import mysql.connector

#User class
class User:
    def __init__(self, username,email,password,dataServerID):
        """
        docstring
        """
        self.username = username
        self.email = email
        self.password = password
        self.dataServerID = dataServerID
    
    def printData(self):
        print(self.username,self.email,self.password,self.dataServerID)

#DataServer Class
class dataServer:
    def __init__(self, serverID=0,maxCapacity=0,ipAddr="",httpPort="",listeningPort="",available=0):
        """
        docstring
        """
        self.serverID = serverID
        self.maxCapacity = maxCapacity
        self.ipAddr = ipAddr
        self.httpPort = httpPort
        self.listeningPort = listeningPort
        self.available = available

#Create a database connection
mydb = mysql.connector.connect(
  host="localhost",
  user="orestis",   
  password="Ore$tis1997",   
  database="myDropbox"
)

mycursor = mydb.cursor()

def getAvailableDataServer():
    sql = "SELECT * FROM DataServers WHERE Available = True"
    mycursor.execute(sql)
    results = mycursor.fetchall()
    numberOfUsers = 0
    server = dataServer()
    for result in results:
        server.serverID = result[0]
        server.maxCapacity = result[1]
        server.ipAddr = result[2]
        server.httpPort=  result[3]
        server.listeningPort = result[4]
        sql2 = "SELECT COUNT(Username) FROM Users WHERE Users.DataServerId = %s"
        val = (server.serverID,)
        mycursor.execute(sql2,val)
        numberOfUsers = mycursor.fetchall()
        break
    #Check if by adding this user the number of users will be equal to maxCapacity
    if numberOfUsers[0][0] + 1 >=  server.maxCapacity:
        #Update availability of the server to false
        sql = "UPDATE DataServers SET Available = False WHERE ServerId = %s"
        val = (server.serverID,)
        mycursor.execute(sql,val)
        mydb.commit()
    return server


#Add a user to the database
def addUser(user):
    sql = "INSERT INTO Users(Username,Email,Password,DataServerId) VALUES(%s,%s,%s,%s) "
    val = (user.username,user.email,user.password,user.dataServerID)
    mycursor.execute(sql,val)
    mydb.commit()

#Get a user from the database
def getUser(username):
    sql = "SELECT * FROM Users WHERE username=%s"
    val = (username,)
    mycursor.execute(sql,val)
    results = mycursor.fetchall()
    if len(results) == 0:
        return None
    user = results[0]
    userInfo = User(user[0],user[1],user[2],user[3])
    return userInfo

#Get dataServer from the database
def getDataServer(serverID):
    sql = "SELECT * FROM DataServers WHERE ServerId = %s"
    val = (serverID,)
    mycursor.execute(sql,val)
    results = mycursor.fetchall()
    server = results[0]
    serverInfo = dataServer(server[0],server[1],server[2],server[3],server[4],server[5])
    return serverInfo