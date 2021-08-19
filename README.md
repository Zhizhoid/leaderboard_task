# leaderboard_task
1) Install Go & MySQL
2) Execute CreateDB.sql file to create database
3) Run SampleValues.go to generate sample values (you can change ValueAmount constant in order to configure sample value amount)
4) Run main.go file to start server
# Authentication
1) Add "Authorization" header to the request (Bearer TOKEN); acces token is "token" ![image](https://user-images.githubusercontent.com/89133139/130067091-faa07c8a-51f8-4148-8408-2a2c0af8ca07.png)
![image](https://user-images.githubusercontent.com/89133139/130067138-4b223aa0-3417-4d4a-b8b3-cc30e027019a.png)
# Requests
1) Store score at http://localhost:1234/leaderboard/store via POST request ![image](https://user-images.githubusercontent.com/89133139/130066193-844b8cee-95b0-434d-bc1f-c4af9f8ca2ee.png)
2) Get score at http://localhost:1234/leaderboard/get via GET request ![image](https://user-images.githubusercontent.com/89133139/130066382-91eb7c28-f44a-454a-8ebd-c69bffd5dccf.png)
