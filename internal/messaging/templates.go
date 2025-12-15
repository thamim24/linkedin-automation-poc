package messaging

// MessageTemplate represents a message template with variables
type MessageTemplate struct {
	Name string
	Body string
}

// FollowUpTemplate is the primary follow-up message template
var FollowUpTemplate = MessageTemplate{
	Name: "follow_up_1",
	Body: "Hi {{first_name}}, great to connect with you! I came across your profile and wanted to say hello.",
}
