package chat

var (
	chat4allReq = `
{
    "action": "next",
    "messages": [
        {
            "id": "aaa2ab9b-548f-460b-8c44-cd4b2030f323",
            "author": {
                "role": "user"
            },
            "content": {
                "content_type": "text",
                "parts": [
                    "这个PDF内容是什么？"
                ]
            },
            "metadata": {
                "attachments": [
                    {
                        "name": "abc2.pdf",
                        "id": "file-N07oWlgbt3rNymA52VaZyw0z",
                        "size": 151618,
                        "mimeType": "application/pdf",
                        "fileTokenSize": 169
                    }
                ]
            }
        }
    ],
    "parent_message_id": "aaa1905b-56c5-4f03-a940-ee6e95bb23fe",
    "model": "gpt-4",
    "timezone_offset_min": -480,
    "suggestions": [
        "What are 5 creative things I could do with my kids' art? I don't want to throw them away, but it's also so much clutter.",
        "Tell me a random fun fact about the Roman Empire",
        "Suggest 5 codenames for a project introducing flexible work arrangements, and the meaning behind each.",
        "Come up with 5 concepts for a retro-style arcade game."
    ],
    "history_and_training_disabled": false,
    "conversation_mode": {
        "kind": "primary_assistant"
    },
    "force_paragen": false,
    "force_rate_limit": false
}
`
	apiError = `
{
	"error": {
		"message": "infomess",
		"type": "openai_hk_error",
		"param": "",
		"code": "openai_hk_error"
	}
}
`
)
