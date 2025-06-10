const express = require('express');
const app = express();
app.use(express.json());

app.post('/', (req, res) => {
  console.log('Echo node request: POST /');
  res.json(req.body);
  console.log('Echo node response sent successfully');
});

app.get('/health', (req, res) => {
  res.status(200).json({ status: 'ok' });
});

app.listen(8082, () => console.log('Node.js echo server started on port 8082'));
