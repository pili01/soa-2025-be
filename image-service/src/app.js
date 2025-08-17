const express = require('express');
const imageRoutes = require('./routes/imageRoutes');
require('dotenv').config();

const app = express();
app.use(express.json());

app.use('/pic', express.static('uploads/pictures'));
app.use('/api', imageRoutes);

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => console.log(`Images service running on port ${PORT}`));