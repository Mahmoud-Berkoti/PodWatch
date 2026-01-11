package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/podwatch/podwatch/pkg/models"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	natsConn *nats.Conn
	// Cache: containerID -> Pod
	// We use a simple map with RWMutex for this MVP as informers keep it up to date.
	// But completed pods might be deleted. We need to keep them for a bit?
	// The informer cache is good enough for active pods.
	// We might need a separate LRU for recently deleted pods if events come in late.
	// For now, let's rely on the informer cache and a secondary map for quick lookup.
	containerToPod sync.Map // map[string]*v1.Pod
)

func main() {
	// 1. NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	var err error
	for i := 0; i < 5; i++ {
		natsConn, err = nats.Connect(natsURL)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer natsConn.Close()

	// 2. K8s Client
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = os.Getenv("HOME") + "/.kube/config"
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalf("Error building kubeconfig: %v", err)
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error building k8s client: %v", err)
	}

	// 3. Informers
	factory := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	podInformer := factory.Core().V1().Pods().Informer()

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			updateCache(obj.(*v1.Pod))
		},
		UpdateFunc: func(old, new interface{}) {
			updateCache(new.(*v1.Pod))
		},
		DeleteFunc: func(obj interface{}) {
			// In a real system we'd keep it for a bit.
			// deleteCache(obj.(*v1.Pod))
		},
	})

	stopCh := make(chan struct{})
	defer close(stopCh)
	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)

	log.Println("Enrich service started, listening for events...")

	// 4. NATS Subscribe
	// Queue group "enrich-workers" ensures load balancing if we run multiple replicas
	_, err = natsConn.QueueSubscribe("events.raw.>", "enrich-workers", func(msg *nats.Msg) {
		enrichEvent(msg)
	})
	if err != nil {
		log.Fatalf("Error subscribing: %v", err)
	}

	select {}
}

func updateCache(pod *v1.Pod) {
	for _, status := range pod.Status.ContainerStatuses {
		// ID is usually "docker://..." or "containerd://..."
		// We strip the prefix or store as is.
		// Falco usually sends the short ID or full ID.
		// Let's store both the full ID and the short ID (12 chars) just in case.
		id := status.ContainerID
		if id != "" {
			// Extract ID part
			parts := strings.Split(id, "://")
			if len(parts) > 1 {
				containerToPod.Store(parts[1], pod)
				if len(parts[1]) > 12 {
					containerToPod.Store(parts[1][:12], pod)
				}
			}
			containerToPod.Store(id, pod)
		}
	}
}

func enrichEvent(msg *nats.Msg) {
	var event models.RuntimeEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	// We only care if there is a container ID
	if event.Container != nil && event.Container.ContainerID != "" {
		// Look up pod
		// Falco container ID is usually short 12 chars
		if podRaw, ok := containerToPod.Load(event.Container.ContainerID); ok {
			pod := podRaw.(*v1.Pod)
			event.Container.Pod = pod.Name
			event.Container.Namespace = pod.Namespace
			event.Container.ServiceAccount = pod.Spec.ServiceAccountName
			event.Container.Labels = pod.Labels

			// If image digest is missing, we might find it in status
			// But Falco usually provides it.
		}
	}

	// Publish to enriched stream
	enrichedData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshalling enriched event: %v", err)
		return
	}

	if err := natsConn.Publish("events.enriched", enrichedData); err != nil {
		log.Printf("Error publishing enriched event: %v", err)
	}
}
