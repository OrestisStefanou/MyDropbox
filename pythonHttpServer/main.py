from flask import Flask,redirect,url_for
from flask import render_template,request,session
from flask import send_file
import json
import socket
import database

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
                return "<h1>Wrong password"
                #SEND TO ERROR PAGE
            session["username"] = userInfo.username
            session["dataServerID"] = userInfo.dataServerID
            return render_template("welcomePage.html",Username=username)
        else:
            return "<h1>Wrong username</h1>"
            #SEND TO ERROR PAGE
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
            return "<h1>Passwords do not match</h1>"
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
            if data["Rtype"] == "OK":
                #Create a user in the database\
                database.addUser(userInfo)
                #Send them to loginPage
                session["username"] = userInfo.username
                session["dataServerID"] = userInfo.dataServerID
                return render_template("welcomePage.html",Username=userInfo.username)
            
        else:
            return "<h1>A username with this username already exists!</h1>"
            #Send to error Page
    else:
        return render_template("signUp.html")

@app.route("/logout")
def logout():
    session.pop("username",None)
    session.pop("dataServerID",None)
    return redirect(url_for("home"))

@app.route("/download")
def download():
    path = "./databaseCopy.sql"
    return send_file(path, as_attachment=True)

@app.route("/user/<name>")
def user(name):
    return f"<h1>{name}</h1>"

if __name__ == "__main__":
    app.run(debug=True)