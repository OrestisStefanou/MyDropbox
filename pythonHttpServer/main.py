from flask import Flask,redirect,url_for
from flask import render_template,request,session
from flask import send_file
import json
import socket
import database

app = Flask(__name__)

@app.route("/home")
@app.route("/")
def home():
    return render_template("index.html")

@app.route("/signup",methods=["POST","GET"])
def signup():
    if request.method == "POST":
        #Handle the form
        username = request.form['entered_username']
        email = request.form['entered_email']
        password = request.form['entered_pass']
        password2 = request.form['entered_pass2']
        #Check if passwords match
        if password != password2:
            print("Passwords do not match")
            #SEND TO AN ERROR PAGE
        #Check if a user with this username already exists
        if database.getUser(username) == None :
            #Create a user
            serverInfo = database.getAvailableDataServer()
            userInfo = database.User(username,email,password,serverInfo.serverID)
            #Send a request to DataServer to create a new User
            s = socket.socket(socket.AF_INET,socket.SOCK_STREAM)
            host = serverInfo.ipAddr
            port = int(serverInfo.listeningPort)
            s.connect((host,port))
            msg = {"From":"HttpServer","Rtype":"createUser","Data":userInfo.username}
            req = json.dumps(msg)
            req = req + '\n'
            s.send(req.encode())
            response = s.recv(1024).decode()
            data = json.loads(response)
            print(data)
            #Check the response and do some shit with it
        else:
            print("A username with this username already exists!")
            #Send to error Page
    else:
        return render_template("signUp.html")

if __name__ == "__main__":
    app.run(debug=True)