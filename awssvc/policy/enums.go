package policy

// Effect describes the effect a policy statement will have upon the resource and for the actions described
type Effect string

var (

	// Allow to grant access of the resource and actions to the principals described in the policy statement
	Allow = Effect("Allow")

	// Deny to deny access of the resource and actions from the principals described in the policy statement
	Deny = Effect("Deny")
)

// Action describes a valid operation that may be made against a particular AWS resource
type Action string

// Describes the various action types available to AWS
var (

	// SQS actions
	AddPermission              = Action("sqs:AddPermission")
	ChangeMessageVisibility    = Action("sqs:ChangeMessageVisibility")
	CreateQueue                = Action("sqs:CreateQueue")
	DeleteMessage              = Action("sqs:DeleteMessage")
	DeleteQueue                = Action("sqs:DeleteQueue")
	GetQueueAttributes         = Action("sqs:GetQueueAttributes")
	GetQueueUrl                = Action("sqs:GetQueueUrl")
	ListDeadLetterSourceQueues = Action("sqs:ListDeadLetterSourceQueues")
	ListQueueTags              = Action("sqs:ListQueueTags")
	ListQueues                 = Action("sqs:ListQueues")
	PurgeQueue                 = Action("sqs:PurgeQueue")
	ReceiveMessage             = Action("sqs:ReceiveMessage")
	RemovePermission           = Action("sqs:RemovePermission")
	SendMessage                = Action("sqs:SendMessage")
	SetQueueAttributes         = Action("sqs:SetQueueAttributes")
	TagQueue                   = Action("sqs:TagQueue")
	UntagQueue                 = Action("sqs:UntagQueue")
	SqsAll                     = Action("sqs:*")

	// KMS actions
	CancelKeyDeletion                   = Action("kms:CancelKeyDeletion")
	ConnectCustomKeyStore               = Action("kms:ConnectCustomKeyStore")
	CreateAlias                         = Action("kms:CreateAlias")
	CreateCustomKeyStore                = Action("kms:CreateCustomKeyStore")
	CreateGrant                         = Action("kms:CreateGrant")
	CreateKey                           = Action("kms:CreateKey")
	Decrypt                             = Action("kms:Decrypt")
	DeleteAlias                         = Action("kms:DeleteAlias")
	DeleteCustomKeyStore                = Action("kms:DeleteCustomKeyStore")
	DeleteImportedKeyMaterial           = Action("kms:DeleteImportedKeyMaterial")
	DescribeCustomKeyStores             = Action("kms:DescribeCustomKeyStores")
	DescribeKey                         = Action("kms:DescribeKey")
	DisableKey                          = Action("kms:DisableKey")
	DisableKeyRotation                  = Action("kms:DisableKeyRotation")
	DisconnectCustomKeyStore            = Action("kms:DisconnectCustomKeyStore")
	EnableKey                           = Action("kms:EnableKey")
	EnableKeyRotation                   = Action("kms:EnableKeyRotation")
	Encrypt                             = Action("kms:Encrypt")
	GenerateDataKey                     = Action("kms:GenerateDataKey")
	GenerateDataKeyPair                 = Action("kms:GenerateDataKeyPair")
	GenerateDataKeyPairWithoutPlaintext = Action("kms:GenerateDataKeyPairWithoutPlaintext")
	GenerateDataKeyWithoutPlaintext     = Action("kms:GenerateDataKeyWithoutPlaintext")
	GenerateMac                         = Action("kms:GenerateMac")
	GenerateRandom                      = Action("kms:GenerateRandom")
	GetKeyPolicy                        = Action("kms:GetKeyPolicy")
	GetKeyRotationStatus                = Action("kms:GetKeyRotationStatus")
	GetParametersForImport              = Action("kms:GetParametersForImport")
	GetPublicKey                        = Action("kms:GetPublicKey")
	ImportKeyMaterial                   = Action("kms:ImportKeyMaterial")
	ListAliases                         = Action("kms:ListAliases")
	ListGrants                          = Action("kms:ListGrants")
	ListKeyPolicies                     = Action("kms:ListKeyPolicies")
	ListKeys                            = Action("kms:ListKeys")
	ListResourceTags                    = Action("kms:ListResourceTags")
	ListRetirableGrants                 = Action("kms:ListRetirableGrants")
	PutKeyPolicy                        = Action("kms:PutKeyPolicy")
	ReEncryptFrom                       = Action("kms:ReEncryptFrom")
	ReEncryptTo                         = Action("kms:ReEncryptTo")
	ReplicateKey                        = Action("kms:ReplicateKey")
	RetireGrant                         = Action("kms:RetireGrant")
	RevokeGrant                         = Action("kms:RevokeGrant")
	ScheduleKeyDeletion                 = Action("kms:ScheduleKeyDeletion")
	Sign                                = Action("kms:Sign")
	TagResource                         = Action("kms:TagResource")
	UntagResource                       = Action("kms:UntagResource")
	UpdateAlias                         = Action("kms:UpdateAlias")
	UpdateCustomKeyStore                = Action("kms:UpdateCustomKeyStore")
	UpdateKeyDescription                = Action("kms:UpdateKeyDescription")
	UpdatePrimaryRegion                 = Action("kms:UpdatePrimaryRegion")
	Verify                              = Action("kms:Verify")
	VerifyMac                           = Action("kms:VerifyMac")
	KmsAll                              = Action("kms:*")
)
