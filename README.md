# Alle-ssment
This project gives a GO API for conversation, image retrival and image download

Swagger API description is present in ./API.yaml
https://editor.swagger.io/#


## Architecture -

Example variables are present in keys-example.sh, edit them and rename file keys.sh

### Backend -
GO based, uses OpenAI package for different role messages.

4 endpoints - 
- /converse     Converse and infer intent
- /upload       Upload image on server
- /download     Fetch image
- /list         List latest images

Run using -
```
source keys.sh
cd src
go run main.go
```

### Frontend -
Made with Python streamlit.

You can hold long running conversations with it.
Ask it to upload and download image. All image have a description which acts as an ID to store and fetch image.

```
source keys.sh
pip install streamlit
cd src
streamlit run ui.py
```