db = db.getSiblingDB(process.env.MONGO_INITDB_DATABASE);

console.log(`Checking for 'tours' collection in database: ${process.env.MONGO_INITDB_DATABASE}`);

const collectionNames = db.getCollectionNames();

if (collectionNames.includes('tours')) {
    console.log("'tours' collection already exists. Skipping creation.");
} else {
    console.log("'tours' collection does not exist. Creating now...");
    db.createCollection('tours', {
      validator: {
        $jsonSchema: {
          bsonType: "object",
          required: ["_id", "authorId", "name", "description", "difficulty", "status", "price"],
          properties: {
            _id: {
              bsonType: "int",
              description: "must be an integer and is required. Managed by the application."
            },
            authorId: {
              bsonType: "int",
              description: "must be an integer and is required."
            },
            name: {
              bsonType: "string",
              description: "must be a string and is required."
            },
            description: {
              bsonType: "string",
              description: "must be a string and is required."
            },
            difficulty: {
              enum: ["Easy", "Medium", "Hard"],
              description: "can only be one of the enum values and is required."
            },
            tags: {
              bsonType: "array",
              description: "must be an array of strings if it exists.",
              items: {
                bsonType: "string"
              }
            },
            status: {
              enum: ["Draft", "Published", "Archived"],
              description: "can only be one of the enum values and is required."
            },
            price: {
              bsonType: "double",
              description: "must be a double and is required."
            }
          }
        }
      }
    });

    console.log("Collection 'tours' created with validation rules.");

    db.tours.createIndex({ "authorId": 1 });
    db.tours.createIndex({ "authorId": 1, "name": 1 }, { unique: true });

    console.log("Indexes for 'tours' collection created/ensured.");
}


console.log(`Checking for 'keypoints' collection in database: ${process.env.MONGO_INITDB_DATABASE}`);

if (collectionNames.includes('keypoints')) {
    console.log("'keypoints' collection already exists. Skipping creation.");
} else {
    console.log("'keypoints' collection does not exist. Creating now...");
    db.createCollection('keypoints', {
      validator: {
        $jsonSchema: {
          bsonType: "object",
          required: ["_id", "tourId", "name", "latitude", "longitude", "ordinal"],
          properties: {
            _id: {
              bsonType: "int",
              description: "must be an integer and is required. Managed by the application."
            },
            tourId: {
              bsonType: "int",
              description: "must be an integer and is required. References the tour this keypoint belongs to."
            },
            name: {
              bsonType: "string",
              description: "must be a string and is required."
            },
            description: {
              bsonType: "string",
              description: "must be a string if it exists."
            },
            imageUrl: {
              bsonType: "string",
              description: "must be a string if it exists."
            },
            latitude: {
              bsonType: "double",
              description: "must be a double and is required. Geographic latitude coordinate."
            },
            longitude: {
              bsonType: "double",
              description: "must be a double and is required. Geographic longitude coordinate."
            },
            ordinal: {
              bsonType: "int",
              description: "must be an integer and is required. Order of the keypoint in the tour."
            }
          }
        }
      }
    });

    console.log("Collection 'keypoints' created with validation rules.");

    
    db.keypoints.createIndex({ "tourId": 1 });
    db.keypoints.createIndex({ "tourId": 1, "ordinal": 1 });
    db.keypoints.createIndex({ "latitude": 1, "longitude": 1 });

    console.log("Indexes for 'keypoints' collection created/ensured.");
}

if (collectionNames.includes('tourExecution')) {
    console.log("'tourExecution' collection already exists. Skipping creation.");
} else {
    console.log("'tourExecution' collection does not exist. Creating now...");
    db.createCollection('tourExecution', {
      validator: {
        $jsonSchema: {
          bsonType: "object",
          required: ["_id", "tour_id", "user_id", "started_at", "status", "last_activity"],
          properties: {
            _id: {
              bsonType: "int",
              description: "must be an integer and is required. Managed by the application."
            },
            tour_id: {
              bsonType: "int",
              description: "must be an integer and is required."
            },
            user_id: {
              bsonType: "int",
              description: "must be an integer and is required."
            },
            started_at: {
              bsonType: "date",
              description: "must be a date and is required."
            },
            ended_at: {
              bsonType: "date",
              description: "must be a date."
            },
            last_activity: {
              bsonType: "date",
              description: "must be a date and is required."
            },
            status: {
              enum: ["pending", "in_progress", "completed", "failed"],
              description: "can only be one of the enum values and is required."
            },
            finished_keypoints: {
              bsonType: "array",
              description: "must be an array of finished keypoints.",
              items: {
                bsonType: "object",
                required: ["keypoint_id", "completed_at"],
                properties: {
                  keypoint_id: {
                    bsonType: "int",
                    description: "must be an integer and is required."
                  },
                  completed_at: {
                    bsonType: "date",
                    description: "must be a date and is required."
                  }
                }
              }
            }
          }
        }
      }
    });

    console.log("Collection 'tourExecution' created with validation rules.");

    db.tourExecution.createIndex({ "tour_id": 1, "user_id": 1 }, { unique: true });

    console.log("Indexes for 'tourExecution' collection created/ensured.");
}

console.log("Database initialization script finished.");