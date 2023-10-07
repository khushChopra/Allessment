This is API doc for Alle assessment -
It involves building an AI chat bot to have conversation and allow it to save and retrive images.

/converse

Takes complete conversation for a user and find the intent,
if intent in save or download image, give key "intent" as "save_image" or "download_image"
else give the next dialogue in conversation

Receive data = 
{
    "msg" : "",
    "history": [
        {
            "role": "user",
            "msg": "msg"
        },
        {
            "role": "assistant",
            "msg": "msg"
        }
    ]
}

Response -
{
    "intent": "download_image/upload_image/None",
    "msg": "Return msg"
}

/upload

Takes description as input and image
Saved the image with a unique ID and stores the image to description mapping in memory or as a file

/download 

Takes description as input
Best matching description for an image is giving as result
