const express = require('express');
const multer = require('multer');
const fs = require('fs');
const path = require('path');
const { v4: uuidv4 } = require('uuid');

const router = express.Router();
const uploadDir = path.join(__dirname, '../../uploads/ProfilePictures');
const reviewUploadDir = path.join(__dirname, '../../uploads/TourReviewPictures');
//const uploadDir = path.join(__dirname, '../../uploads/images');

if (!fs.existsSync(uploadDir)) fs.mkdirSync(uploadDir, { recursive: true });
if (!fs.existsSync(reviewUploadDir)) fs.mkdirSync(reviewUploadDir, { recursive: true });

const upload = multer({ storage: multer.memoryStorage() });

router.post('/save-image', upload.single('image'), (req, res) => {
    if (!req.file) {
        return res.status(400).json({ error: 'No file uploaded' });
    }

    const uniqueId = uuidv4();
    const ext = path.extname(req.file.originalname);
    const filename = `${uniqueId}-${Date.now()}${ext}`;
    const filePath = path.join(uploadDir, filename);

    fs.writeFileSync(filePath, req.file.buffer);

    const baseURL = process.env.BASE_UPLOAD_URL;

    res.status(201).json({
        message: 'Image saved successfully',
        photoURL: `${baseURL}/api/img/${filename}`
    });
});

router.get('/img/:filename', (req, res) => {
    const filename = req.params.filename;
    const filePath = path.join(uploadDir, filename);

    if (!fs.existsSync(filePath)) {
        return res.status(404).json({ error: 'File not found' });
    }

    res.sendFile(filePath);
});


router.post('/saveReviewPhoto', upload.single('image'), (req, res) => {
    if (!req.file) {
        return res.status(400).json({ error: 'No file uploaded' });
    }

    const userId = req.body.userId || 'unknown_user';
    const tourId = req.body.tourId || 'unknown_tour';
    const ext = path.extname(req.file.originalname);

    const filename = `tour-${tourId}-user-${userId}-${Date.now()}${ext}`;
    const filePath = path.join(reviewUploadDir, filename);

    fs.writeFileSync(filePath, req.file.buffer);

    res.status(201).json({
        message: 'Review image saved successfully',
        photoURL: `http://localhost:3031/api/images/review/${filename}`  
    });
});

router.get('/images/review/:filename', (req, res) => {
    const filename = req.params.filename;
    const filePath = path.join(reviewUploadDir, filename);

    if (!fs.existsSync(filePath)) {
        return res.status(404).json({ error: 'File not found' });
    }
    
    res.sendFile(filePath);
});

module.exports = router;