const express = require('express');
require('dotenv').config();

const distanceRoutes = require('./routes/distanceRoutes');

const app = express();
app.use(express.json());

app.use(distanceRoutes);

app.get('/', (req, res) => {
  res.send('Map Service using OpenRouteService is running');
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Map service running on port ${PORT}`);
});

module.exports = app;