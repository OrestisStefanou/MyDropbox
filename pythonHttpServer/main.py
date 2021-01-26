from flask import Flask,redirect,url_for
from flask import render_template,request,session
from flask import send_file
import json
import socket
import database
import os

app = Flask(__name__)
app.secret_key = "secretKey"

@app.route("/home")
@app.route("/")
def home():
    if "username" in session:
        return render_template("welcomePage.html",Username = session["username"])
    return render_template("index.html")

@app.route("/signin",methods=["POST","GET"])
def signin():
    if "username" in session:
        return render_template("welcomePage.html",Username = session["username"])
    if request.method == "POST":
        username = request.form['entered_username']
        password = request.form['entered_pass']
        #Check if there is a user with this username
        userInfo = database.getUser(username)
        if userInfo:
            if password != userInfo.password:
                return render_template("errorPage.html",errorMessage = "Wrong password")
            session["username"] = userInfo.username
            session["dataServerID"] = userInfo.dataServerID
            serverInfo = database.getDataServer(userInfo.dataServerID)
            #Send a request to dataServer to get listening port of user's file server
            s = socket.socket(socket.AF_INET,socket.SOCK_STREAM)
            host = serverInfo.ipAddr
            port = int(serverInfo.listeningPort)
            s.connect((host,port))
            msg = {"From":"HttpServer","Rtype":"UserLogin","Data":userInfo.username}
            req = json.dumps(msg)
            req = req + '\n'
            s.send(req.encode())
            response = s.recv(1024).decode()
            s.close()
            data = json.loads(response)
            print(data)
            fileServerPort = data["Data"]
            if data["Rtype"] == "Error":
                #Send error html page
                return render_template("errorPage.html",errorMessage = 'Internal server error')
            
            fileServerUrl = "http://{}:{}".format(serverInfo.ipAddr,fileServerPort)
            return render_template("welcomePage.html",Username=username,fileServerURL = fileServerUrl)
        else:
            return render_template("errorPage.html",errorMessage="No user with this username")
    else:
        return render_template("signIn.html")

@app.route("/signup",methods=["POST","GET"])
def signup():
    if "username" in session:
        return render_template("welcomePage.html",Username = session["username"])
    if request.method == "POST":
        #Handle the form
        username = request.form['entered_username']
        email = request.form['entered_email']
        password = request.form['entered_pass']
        password2 = request.form['entered_pass2']
        #Check if passwords match
        if password != password2:
            #Send user to error page
            return render_template("errorPage.html",errorMessage = "Passwords do not match")
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
            s.close()
            print(data)
            if data["Rtype"] == "OK":
                #Create a user in the database\
                database.addUser(userInfo)
                #Send them to loginPage
                session["username"] = userInfo.username
                session["dataServerID"] = userInfo.dataServerID
                fileServerPort = data["Data"]
                fileServerUrl = "http://{}:{}".format(serverInfo.ipAddr,fileServerPort)
                print(fileServerUrl)
                return render_template("welcomePage.html",Username=userInfo.username,fileServerURL = fileServerUrl)
            
        else:
            #Send user to error page
            return render_template("errorPage.html",errorMessage = "A username with this username already exists!")
    else:
        return render_template("signUp.html")

@app.route("/editProfile",methods=["POST","GET"])
def editProfile():
    if request.method == "POST":
        #Handle the form
        email = request.form['entered_email']
        password = request.form['entered_pass']
        password2 = request.form['entered_pass2']
        #Check if passwords match
        if password != password2:
            #Send user to error page
            return render_template("errorPage.html",errorMessage = "Passwords do not match")
        database.updateUserInfo(session["username"],password,email)
        return render_template("welcomePage.html",Username = session["username"])
    else:
        return render_template("editProfile.html")   

@app.route("/logout")
def logout():
    session.pop("username",None)
    session.pop("dataServerID",None)
    return redirect(url_for("home"))

@app.route("/download")
def download():
    path = "./databaseCopy.sql"
    return send_file(path, as_attachment=True)

@app.route("/<user>",methods=["POST","GET"])
def user(user):
    if request.method == "POST":
        username = str(request.form["key"])
        return request.data
    baseDir = "/home/orestis/MyDropboxClients"
    userDir = os.path.join(baseDir,user)
    try:
        userFiles = os.scandir(userDir)
    except:
        return "No user "
    for userFile in userFiles:
        print(userFile.name,userFile.is_dir())
    return render_template("welcomePage.html",Username=user,fileServerURL = fileServerUrl)

if __name__ == "__main__":
    app.run(debug=True)