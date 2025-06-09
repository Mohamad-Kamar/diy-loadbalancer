const express = require('express');
const app = express();
app.use(express.json());

app.post('/', (req, res) => {
  res.json(req.body);
});

app.get('/health', (req, res) => {
  res.status(200).json({ status: 'ok' });
});

app.listen(8082, () => console.log('Echo Node listening on 8082'));
