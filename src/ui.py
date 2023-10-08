import streamlit as st
import requests, re, os
URL = os.environ.get("SERVER_ENDPOINT")

def make_converse_request(prompt, history):
    url = URL+"/converse"
    headers = {"Content-Type": "application/json"}
    data = [
        {
            "role": x["role"],
            "msg": x["content"]
        }
        for x in history if x["content"] not in ["Found file", "Received file"] or x["role"]!="image"
    ]
    data = {
        "history": data,
        "msg": prompt
    }
    try:

        response = requests.post(url, json=data, headers=headers)
        if response.status_code == 200:
            result = response.json()
            return result
    except Exception as e:
        pass
    return {"intent": "", "msg": f"Error"}

def make_upload_request(fname, fdata, description):
    url = URL+"/upload"
    data = {
        "description": description
    }
    files = {
        "image": (fname, fdata)
    }
    try:
        response = requests.post(url, data=data, files=files)
        if response.status_code == 200:
            return 1
    except Exception as e:
        pass
    return 0

def make_download_request(description):
    url = URL+"/download?description="+description
    try:
        response = requests.get(url)
        if response.status_code == 200:
            d = response.headers['content-disposition']
            fname = re.findall("filename=(.+)", d)[0]
            return response.content, fname
    except Exception as e:
        pass
    return None

st.title("Alle-ssment")
if "messages" not in st.session_state:
    st.session_state.messages = []
if "state" not in st.session_state:
    st.session_state.state = 0
for message in st.session_state.messages:
    if message["role"]=="image":
        with st.chat_message("assistant"):
            st.image(message["content"])
    else:
        with st.chat_message(message["role"]):
            st.markdown(message["content"])

if st.session_state.state==1:
    uploaded_file = st.file_uploader("Choose an image")
    description = st.text_input("Image identifier")
    if uploaded_file is not None:
        bytes_data = uploaded_file.getvalue()
    if st.button("Upload!"):
        make_upload_request(uploaded_file.name, uploaded_file.getvalue(), description)
        st.session_state.state = 0
        st.session_state.messages.append({"role": "assistant", "content": "Received file"})
        st.rerun()

if st.session_state.state==2:
    description = st.text_input("Image identifier")
    if st.button("Get image"):
        data, name = make_download_request(description)
        st.session_state.messages.append({"role": "image", "content": data})
        st.session_state.state = 0
        st.rerun()

if prompt := st.chat_input("What is up?", disabled=st.session_state.state!=0):
    resp = make_converse_request(prompt, st.session_state.messages)
    intent = resp["intent"]
    if intent=="upload_image":
        st.session_state.state = 1
        st.rerun()
    elif intent=="download_image":
        st.session_state.state = 2
        st.rerun()
    else:
        with st.chat_message("user"):
            st.markdown(prompt)
        st.session_state.messages.append({"role": "user", "content": prompt})
        with st.chat_message("assistant"):
            st.markdown(resp["msg"])
        st.session_state.messages.append({"role": "assistant", "content": resp["msg"]})