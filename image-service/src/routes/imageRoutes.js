const express = require('express');
const multer = require('multer');
const fs = require('fs');
const path = require('path');

const router = express.Router();
const uploadDir = path.join(__dirname, '../../uploads/ProfilePictures');

if (!fs.existsSync(uploadDir)) fs.mkdirSync(uploadDir, { recursive: true });

const upload = multer({ storage: multer.memoryStorage() });

router.post('/saveProfilePhoto', upload.single('image'), (req, res) => {
    if (!req.file) {
        return res.status(400).json({ error: 'No file uploaded' });
    }

    const userId = req.body.userId || 'unknown';
    const ext = path.extname(req.file.originalname);
    const filename = `${userId}-${Date.now()}${ext}`;
    const filePath = path.join(uploadDir, filename);

    fs.writeFileSync(filePath, req.file.buffer);

    res.status(201).json({
        message: 'Image saved successfully',
        photoURL: `http://localhost:3031/api/profilePhoto/${filename}`  
    });
});

router.get('/profilePhoto/:filename', (req, res) => {
    const filename = req.params.filename;
    const filePath = path.join(uploadDir, filename);

    if (!fs.existsSync(filePath)) {
        return res.status(404).json({ error: 'File not found' });
    }

    res.sendFile(filePath);
});

module.exports = router;