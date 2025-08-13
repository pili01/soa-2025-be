 require('dotenv').config(); 

const express = require('express');
const apiRouter = require('./api/routes');

const app = express();
const port = process.env.PORT || 3000;

app.use(express.json());

app.use('/api', apiRouter);

app.use((req, res, next) => {
  res.status(404).json({
    success: false,
    message: 'Tražena ruta nije pronađena na ovom serveru.',
  });
});

app.listen(port, () => {
  console.log(`Blog servis sluša na portu ${port}`);
});

module.exports = app;