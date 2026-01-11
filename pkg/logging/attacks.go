package logging

// Attack Types - Common container/K8s attack patterns
const (
	AttackShellSpawn          = "shell_spawn"
	AttackTokenTheft          = "token_theft"
	AttackPrivilegeEscalation = "privilege_escalation"
	AttackReverseShell        = "reverse_shell"
	AttackLateralMovement     = "lateral_movement"
	AttackDataExfiltration    = "data_exfiltration"
	AttackCryptoMining        = "crypto_mining"
	AttackContainerEscape     = "container_escape"
	AttackResourceHijacking   = "resource_hijacking"
	AttackCredentialAccess    = "credential_access"
	AttackPersistence         = "persistence"
	AttackDefenseEvasion      = "defense_evasion"
	AttackDiscovery           = "discovery"
	AttackPackageInstall      = "package_install"
)

// MITRE ATT&CK Tactics (Containers)
const (
	TacticInitialAccess    = "TA0001"
	TacticExecution        = "TA0002"
	TacticPersistence      = "TA0003"
	TacticPrivilegeEsc     = "TA0004"
	TacticDefenseEvasion   = "TA0005"
	TacticCredentialAccess = "TA0006"
	TacticDiscovery        = "TA0007"
	TacticLateralMovement  = "TA0008"
	TacticCollection       = "TA0009"
	TacticExfiltration     = "TA0010"
	TacticImpact           = "TA0040"
)

// MITRE ATT&CK Techniques (Container-specific)
const (
	TechniqueExecInContainer     = "T1609"     // Container Administration Command
	TechniqueEscapeToHost        = "T1611"     // Escape to Host
	TechniqueContainerAPI        = "T1552.007" // Container API
	TechniquePrivilegedContainer = "T1610"     // Deploy Container
	TechniqueImplantContainer    = "T1525"     // Implant Internal Image
	TechniqueKubeAPI             = "T1552.004" // Unsecured Credentials: Kubernetes Secrets
)

// Kill Chain Phases
const (
	KillChainReconnaissance = "reconnaissance"
	KillChainWeaponization  = "weaponization"
	KillChainDelivery       = "delivery"
	KillChainExploitation   = "exploitation"
	KillChainInstallation   = "installation"
	KillChainC2             = "command_and_control"
	KillChainActions        = "actions_on_objectives"
)

// Severity Levels
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
	SeverityInfo     = "info"
)

// Response Actions
const (
	ResponseKillPod           = "kill_pod"
	ResponseQuarantineNS      = "quarantine_namespace"
	ResponseIsolateNode       = "isolate_node"
	ResponseEvidenceBundle    = "evidence_bundle"
	ResponseNotify            = "notify"
	ResponseBlockNetwork      = "block_network"
	ResponseRevokeCredentials = "revoke_credentials"
)

// Response Status
const (
	StatusPending   = "pending"
	StatusExecuting = "executing"
	StatusSuccess   = "success"
	StatusFailed    = "failed"
	StatusBlocked   = "blocked"
	StatusSkipped   = "skipped"
)

// AttackMapping provides MITRE ATT&CK mappings for common attacks
var AttackMapping = map[string]struct {
	Technique       string
	Tactic          string
	KillChainPhase  string
	DefaultSeverity string
}{
	AttackShellSpawn: {
		Technique:       TechniqueExecInContainer,
		Tactic:          TacticExecution,
		KillChainPhase:  KillChainExploitation,
		DefaultSeverity: SeverityHigh,
	},
	AttackTokenTheft: {
		Technique:       TechniqueKubeAPI,
		Tactic:          TacticCredentialAccess,
		KillChainPhase:  KillChainActions,
		DefaultSeverity: SeverityHigh,
	},
	AttackPrivilegeEscalation: {
		Technique:       TechniqueEscapeToHost,
		Tactic:          TacticPrivilegeEsc,
		KillChainPhase:  KillChainExploitation,
		DefaultSeverity: SeverityCritical,
	},
	AttackReverseShell: {
		Technique:       TechniqueExecInContainer,
		Tactic:          TacticExecution,
		KillChainPhase:  KillChainC2,
		DefaultSeverity: SeverityCritical,
	},
	AttackContainerEscape: {
		Technique:       TechniqueEscapeToHost,
		Tactic:          TacticPrivilegeEsc,
		KillChainPhase:  KillChainExploitation,
		DefaultSeverity: SeverityCritical,
	},
	AttackCryptoMining: {
		Technique:       TechniqueExecInContainer,
		Tactic:          TacticImpact,
		KillChainPhase:  KillChainActions,
		DefaultSeverity: SeverityHigh,
	},
	AttackPackageInstall: {
		Technique:       TechniqueExecInContainer,
		Tactic:          TacticExecution,
		KillChainPhase:  KillChainInstallation,
		DefaultSeverity: SeverityMedium,
	},
}

// GetAttackContext creates an AttackContext from a known attack type
func GetAttackContext(attackType, ruleName, ruleID string, indicators []string) *AttackContext {
	mapping, ok := AttackMapping[attackType]
	if !ok {
		return &AttackContext{
			Type:       attackType,
			RuleName:   ruleName,
			RuleID:     ruleID,
			Indicators: indicators,
			Severity:   SeverityMedium,
			Confidence: 0.5,
		}
	}

	return &AttackContext{
		Type:           attackType,
		Technique:      mapping.Technique,
		TacticID:       mapping.Tactic,
		KillChainPhase: mapping.KillChainPhase,
		Severity:       mapping.DefaultSeverity,
		Confidence:     0.9,
		RuleName:       ruleName,
		RuleID:         ruleID,
		Indicators:     indicators,
	}
}
