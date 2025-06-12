const express = require('express');
const app = express();
app.use(express.json());

app.post('/', (req, res) => {
  console.log(`[node] Received request: ${req.method} ${req.path}`);
  res.json(req.body);
  console.log('[node] Request process=ed successfully');
});

app.get('/health', (req, res) => {
  res.status(200).json({ status: 'ok' });
});

app.listen(8082, () => console.log('[node] Server started on port 8082'));
