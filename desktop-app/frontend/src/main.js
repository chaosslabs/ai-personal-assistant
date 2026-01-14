import './style.css';
import './app.css';

import { GetVersion } from '../wailsjs/go/main/App';

document.querySelector('#app').innerHTML = `
    <div>
      <div class="result" id="result">AI Personal Assistant</div>
      <div class="result" id="version">Loading version...</div>
    </div>
`;

let versionElement = document.getElementById("version");

// Load and display the version
try {
    GetVersion()
        .then((version) => {
            versionElement.innerText = `Version: ${version}`;
        })
        .catch((err) => {
            versionElement.innerText = "Version: unknown";
            console.error(err);
        });
} catch (err) {
    versionElement.innerText = "Version: unknown";
    console.error(err);
}
