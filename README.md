# ButtonMania - A Fun Telegram Mini App! 🕹️

![ButtonMania](https://github.com/ButtonMania/ButtonMania/raw/main/frontend/src/assets/poster.png "ButtonMania")

🎮 Button Mania is a fun and addictive mini app for Telegram. It consists of a server written in Go using the Gin framework and the Telego library, as well as a client written in TypeScript using the Preact framework. 🚀 Redis works its magic as the main database, storing active sessions, leaderboards, and user records. 💾  

Want to try it out? You can easily deploy this awesome Telegram mini app on your server and even run it with Docker.  It's a hilarious and entertaining game that you can try live at [@ButtonManiaBot](https://t.me/ButtonManiaBot) on Telegram. 🤩  


## Features

- **Server in Golang**: The backend server is written in Golang and uses Gin and Telego for seamless communication with Telegram. 🖥️  

- **Client in TypeScript**: The frontend client is developed using TypeScript and Preact, providing an engaging user interface. 🌟  

- **Docker Ready**: We've included a Docker Compose file for easy deployment and usage. 🐳  

- **Configurable**: You can customize the server with various CLI parameters or environment variables, providing flexibility in your deployment. ⚙️  

- **Localizable**: Easily incorporate additional languages in both the frontend and backend. We store all strings in plain text files, making it a breeze to edit, introduce new languages, and expand for new features. 🌐  


## Getting Started

To run the project, you can use Docker with the provided `docker-compose.yml` file. Make sure to set the necessary environment variables (or CLI parameters) as mentioned below.  
Please note that there are two parameters (`TG_APP_URL` and `TG_BOT_TOKEN`) that are not predefined in the `docker-compose.yml` file, so please pay attention to this. 👩‍💻  

1. Clone the repository:

```sh
git clone https://github.com/ButtonMania/ButtonMania.git
cd ButtonMania
```

2. Build and run the Docker containers:

```sh
docker-compose up --build
```

3. Access the app in your browser at http://localhost:8080.


## Overview

### Frontend
The frontend code is located in the `frontend` folder and contains the Preact app. The frontend app communicates with the server via WebSocket and HTTP.  
Each game session sends the initData (received from the Telegram Mini App Api during launch), along with ButtonType and ButtonPhase, via WebSocket. The server parses and validates the initData, and then uses it to update the internal state of the game session.

### Backend
The backend code is located in the `backend` folder and includes both the web and bot parts. The web part stands for WebSocket and HTTP API.  
There is one REST API method that returns statistics for different ButtonTypes (game rooms). The statistics include the current count of players and the total count of players who have ever played in that room (ButtonType).  

### Server Logic
The server creates rooms for each button type (Love, Peace, Fortune, and Prestige) and waits for incoming WebSocket connections. Users must hold the button, and the client must maintain the connection and periodically send messages with the current ButtonPhase (push, hold, release). The server updates the internal state of the user's game based on the current timestamp when a message is received. If the server receives a message with ButtonPhase equal to 'release', the game session is closed, and the record is written to the leaderboard. During user holds, the server sends client update messages, which may contain funny messages.  
Motivational messages are sent by the server to the client at various frequencies, starting every 5 seconds and slowing down while holding the button. These messages are localizable and stored in `./backend/<locale>/messages/<ButtonType>.txt` files.

### Telegram Bot
The Telegram bot subroutine contains two commands: `/start`, which displays a message and a button to start the app, and `/donate`, which displays a message and cryptocurrency addresses for donations. All messages are localizable and stored in `./backend/<locale>/bot/<command>.txt` files.


## Game Logic

![ButtonMania](https://github.com/ButtonMania/ButtonMania/raw/main/frontend/src/assets/meme.gif "ButtonMania")

ButtonMania is a delightful and straightforward game that operates within Telegram. Users pick a button from three options: love, peace, or fortune. Premium Telegram users can access an additional button called "prestige." The objective is to hold the chosen button for as long as possible. During gameplay, motivational messages may (or may not, since it's ironic) be displayed. The server maintains a leaderboard and records game sessions. However, only the best results are stored, and users can view their own rankings.


## Frontend CLI Commands

For the frontend, you can use the following CLI commands:

- `npm install`: Installs dependencies.
- `npm run dev`: Runs a development, HMR server.
- `npm run serve`: Runs a production-like server.
- `npm run build`: Production-ready build.
- `npm run lint`: Lints TypeScript files using ESLint.
- `npm run test`: Runs Jest and Enzyme tests.

## Server CLI Parameters

Here's a list of CLI parameters and corresponding environment variables that you can use as fallback:

- `redisaddress`: Redis server address (Required). Env: `REDIS_ADDRESS`
- `redisusername`: Redis server username. Env: `REDIS_USERNAME`
- `redispassword`: Redis server password. Env: `REDIS_PASSWORD`
- `redisdatabase`: Redis server database number. Env: `REDIS_DB`
- `redistls`: Redis server TLS connection. Env: `REDIS_TLS`
- `staticpath`: Static assets folder path (Required). Env: `STATIC_PATH`
- `sessionname`: Server session name. Env: `SESSION_NAME`
- `sessionsecret`: Server session secret phrase. Env: `SESSION_SECRET`
- `serverport`: Server port. Env: `SERVER_PORT`
- `allowedorigins`: Allowed CORS origins. Env: `CORS_ORIGINS`
- `telegramappurl`: Telegram app URL (Required). Env: `TG_APP_URL`
- `telegramtoken`: Telegram bot token (Required). Env: `TG_BOT_TOKEN`
- `telegramwebhook`: Telegram webhook URL (if not provided, long polling will be used). Env: `TG_WEBHOOK_URL`
- `telegramwhport`: Telegram webhook listen port. Env: `TG_WEBHOOK_PORT`
- `telegramdonateton`: TON address for Telegram bot donation feature. Env: `TG_DONATION_TON`
- `telegramdonateeth`: Ethereum address for Telegram bot donation feature. Env: `TG_DONATION_ETH`
- `telegramdonatexmr`: Monero address for Telegram bot donation feature. Env: `TG_DONATION_XMR`

**Important Environment Variables:**

- `GIN_MODE`: Controls the debug mode of the server.  
- `CORS_ORIGINS`: Accepts glob patterns and controls allowed CORS origins.  

In debug mode (GIN_MODE=debug), the server accepts requests from all origins, but in release mode (GIN_MODE=release), it only allows hosts listed in CORS_ORIGINS.  


## Localization

ButtonMania supports multiple languages, including English and Russian. Motivational messages are generated on the server side and are located at `./backend/localization/<locale>/messages/<button_type>.txt` files. UI strings are managed on the client side and are located at `./fronted/src/locales/<locale>.json`. 🌍  


## Contributing

ButtonMania is an open-source project, and we welcome contributions from the community. You can help by:

- Adding new localizations.
- Identifying and fixing bugs.
- Writing tests for both the client and server.
- Sharing your ideas and suggestions.

Feel free to open pull requests or issues!

## Donate

If you've enjoyed ButtonMania and would like to support the project, consider making a donation in cryptocurrency. Your contributions help cover hosting costs and other digital expenses. Here are our donation addresses:

- **TON**: `UQAaTJqQ4bqy6xxCUV-MSWMsJulwLAP1Dyma5TaA0aGwWiEe`
- **Ethereum/Binance (ETH/BNB)**: `0x0948A61328b3eCeDa37CC33907F30d4AC06C34Ed`
- **Monero (XMR)**: `48D2unYK1NhfzQusXnXsU6ZrXfPxfSXrKPwTxEknwJygeC6wTSBkWorbX55EYZbBMHZLdeG1GXL8N9Xs6KSFCdEQ5xgoTqg`

Your support helps keep this project alive! 😄  
If you've profited from this code, please consider sending some crypto our way, or reach out to discuss collaboration. We have plenty of ideas and skills!  


## Credits

Programming by Pavel Litvinenko ([@gerasim13](https://github.com/gerasim13))  
Design and Look and Feel by Maria Litvinenko ([@milk010](https://www.linkedin.com/in/milk010))  
Original Idea by Stanislav Khimich ([@BoshaKokosha](https://t.me/BoshaKokosha))  

🚀 Have fun with ButtonMania! 🚀