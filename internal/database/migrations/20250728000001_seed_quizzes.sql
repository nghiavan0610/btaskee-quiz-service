-- +goose Up
-- +goose StatementBegin

-- Insert sample quizzes
INSERT INTO quizzes (title, description, owner_id, visibility, slug, total_questions) VALUES
('General Knowledge Quiz', 'Test your general knowledge with these 20 challenging questions covering various topics including history, science, geography, and culture.', 1, 'published', 'general-knowledge-quiz', 20),
('Science & Technology Quiz', 'Explore the fascinating world of science and technology with these 10 questions about physics, chemistry, biology, and modern tech.', 1, 'published', 'science-technology-quiz', 10),
('History & Geography Quiz', 'Journey through time and around the world with these 10 questions about historical events and geographical facts.', 1, 'published', 'history-geography-quiz', 10);

-- Get the quiz IDs for reference
-- Quiz 1: General Knowledge Quiz (20 questions)
INSERT INTO questions (quiz_id, question, type, answers, time_limit, index) VALUES
-- Question 1
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'What is the capital of Australia?', 
 'single_choice', 
 '[{"text": "Sydney", "is_correct": false}, {"text": "Melbourne", "is_correct": false}, {"text": "Canberra", "is_correct": true}, {"text": "Perth", "is_correct": false}]'::jsonb,
 '20', 0),

-- Question 2
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'Which planet is known as the Red Planet?', 
 'single_choice', 
 '[{"text": "Venus", "is_correct": false}, {"text": "Mars", "is_correct": true}, {"text": "Jupiter", "is_correct": false}, {"text": "Saturn", "is_correct": false}]'::jsonb,
 '20', 1),

-- Question 3
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'Who painted the Mona Lisa?', 
 'single_choice', 
 '[{"text": "Vincent van Gogh", "is_correct": false}, {"text": "Pablo Picasso", "is_correct": false}, {"text": "Leonardo da Vinci", "is_correct": true}, {"text": "Michelangelo", "is_correct": false}]'::jsonb,
 '20', 2),

-- Question 4
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'What is the largest ocean on Earth?', 
 'single_choice', 
 '[{"text": "Atlantic Ocean", "is_correct": false}, {"text": "Indian Ocean", "is_correct": false}, {"text": "Arctic Ocean", "is_correct": false}, {"text": "Pacific Ocean", "is_correct": true}]'::jsonb,
 '20', 3),

-- Question 5
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'Which of the following are programming languages?', 
 'multiple_choice', 
 '[{"text": "Python", "is_correct": true}, {"text": "JavaScript", "is_correct": true}, {"text": "HTML", "is_correct": false}, {"text": "CSS", "is_correct": false}]'::jsonb,
 '45', 4),

-- Question 6
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'In which year did World War II end?', 
 'single_choice', 
 '[{"text": "1944", "is_correct": false}, {"text": "1945", "is_correct": true}, {"text": "1946", "is_correct": false}, {"text": "1947", "is_correct": false}]'::jsonb,
 '20', 5),

-- Question 7
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'What is the chemical symbol for gold?', 
 'text_input', 
 '[{"text": "Au", "is_correct": true}, {"text": "AU", "is_correct": true}, {"text": "au", "is_correct": true}]'::jsonb,
 '20', 6),

-- Question 8
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'Which continent is the Sahara Desert located in?', 
 'single_choice', 
 '[{"text": "Asia", "is_correct": false}, {"text": "Africa", "is_correct": true}, {"text": "Australia", "is_correct": false}, {"text": "South America", "is_correct": false}]'::jsonb,
 '20', 7),

-- Question 9
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'What is the smallest country in the world?', 
 'single_choice', 
 '[{"text": "Monaco", "is_correct": false}, {"text": "San Marino", "is_correct": false}, {"text": "Vatican City", "is_correct": true}, {"text": "Liechtenstein", "is_correct": false}]'::jsonb,
 '20', 8),

-- Question 10
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'Which elements are noble gases?', 
 'multiple_choice', 
 '[{"text": "Helium", "is_correct": true}, {"text": "Neon", "is_correct": true}, {"text": "Oxygen", "is_correct": false}, {"text": "Nitrogen", "is_correct": false}]'::jsonb,
 '45', 9),

-- Question 11
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'Who wrote the novel "1984"?', 
 'single_choice', 
 '[{"text": "Aldous Huxley", "is_correct": false}, {"text": "George Orwell", "is_correct": true}, {"text": "Ray Bradbury", "is_correct": false}, {"text": "H.G. Wells", "is_correct": false}]'::jsonb,
 '20', 10),

-- Question 12
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'What is the currency of Japan?', 
 'text_input', 
 '[{"text": "Yen", "is_correct": true}, {"text": "yen", "is_correct": true}, {"text": "YEN", "is_correct": true}]'::jsonb,
 '20', 11),

-- Question 13
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'Which mountain range contains Mount Everest?', 
 'single_choice', 
 '[{"text": "Andes", "is_correct": false}, {"text": "Himalayas", "is_correct": true}, {"text": "Alps", "is_correct": false}, {"text": "Rocky Mountains", "is_correct": false}]'::jsonb,
 '20', 12),

-- Question 14
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'What are the three primary colors?', 
 'multiple_choice', 
 '[{"text": "Red", "is_correct": true}, {"text": "Blue", "is_correct": true}, {"text": "Yellow", "is_correct": true}, {"text": "Green", "is_correct": false}]'::jsonb,
 '45', 13),

-- Question 15
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'In which city is the famous Colosseum located?', 
 'single_choice', 
 '[{"text": "Athens", "is_correct": false}, {"text": "Rome", "is_correct": true}, {"text": "Paris", "is_correct": false}, {"text": "Madrid", "is_correct": false}]'::jsonb,
 '20', 14),

-- Question 16
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'What is the largest mammal in the world?', 
 'single_choice', 
 '[{"text": "African Elephant", "is_correct": false}, {"text": "Blue Whale", "is_correct": true}, {"text": "Giraffe", "is_correct": false}, {"text": "Sperm Whale", "is_correct": false}]'::jsonb,
 '20', 15),

-- Question 17
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'Which scientist developed the theory of relativity?', 
 'single_choice', 
 '[{"text": "Isaac Newton", "is_correct": false}, {"text": "Albert Einstein", "is_correct": true}, {"text": "Galileo Galilei", "is_correct": false}, {"text": "Stephen Hawking", "is_correct": false}]'::jsonb,
 '20', 16),

-- Question 18
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'What is the freezing point of water in Celsius?', 
 'text_input', 
 '[{"text": "0", "is_correct": true}, {"text": "0°C", "is_correct": true}, {"text": "zero", "is_correct": true}]'::jsonb,
 '20', 17),

-- Question 19
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'Which countries are part of North America?', 
 'multiple_choice', 
 '[{"text": "United States", "is_correct": true}, {"text": "Canada", "is_correct": true}, {"text": "Mexico", "is_correct": true}, {"text": "Brazil", "is_correct": false}]'::jsonb,
 '45', 18),

-- Question 20
((SELECT id FROM quizzes WHERE slug = 'general-knowledge-quiz'), 
 'What is the speed of light in vacuum?', 
 'single_choice', 
 '[{"text": "299,792,458 m/s", "is_correct": true}, {"text": "300,000,000 m/s", "is_correct": false}, {"text": "299,000,000 m/s", "is_correct": false}, {"text": "301,000,000 m/s", "is_correct": false}]'::jsonb,
 '20', 19);

-- Quiz 2: Science & Technology Quiz (10 questions)
INSERT INTO questions (quiz_id, question, type, answers, time_limit, index) VALUES
-- Question 1
((SELECT id FROM quizzes WHERE slug = 'science-technology-quiz'), 
 'What does CPU stand for?', 
 'single_choice', 
 '[{"text": "Central Processing Unit", "is_correct": true}, {"text": "Computer Processing Unit", "is_correct": false}, {"text": "Central Program Unit", "is_correct": false}, {"text": "Computer Program Unit", "is_correct": false}]'::jsonb,
 '20', 0),

-- Question 2
((SELECT id FROM quizzes WHERE slug = 'science-technology-quiz'), 
 'Which of these are renewable energy sources?', 
 'multiple_choice', 
 '[{"text": "Solar", "is_correct": true}, {"text": "Wind", "is_correct": true}, {"text": "Coal", "is_correct": false}, {"text": "Natural Gas", "is_correct": false}]'::jsonb,
 '45', 1),

-- Question 3
((SELECT id FROM quizzes WHERE slug = 'science-technology-quiz'), 
 'What is the chemical formula for water?', 
 'text_input', 
 '[{"text": "H2O", "is_correct": true}, {"text": "h2o", "is_correct": true}]'::jsonb,
 '20', 2),

-- Question 4
((SELECT id FROM quizzes WHERE slug = 'science-technology-quiz'), 
 'Which programming language was developed by Guido van Rossum?', 
 'single_choice', 
 '[{"text": "Java", "is_correct": false}, {"text": "Python", "is_correct": true}, {"text": "C++", "is_correct": false}, {"text": "JavaScript", "is_correct": false}]'::jsonb,
 '20', 3),

-- Question 5
((SELECT id FROM quizzes WHERE slug = 'science-technology-quiz'), 
 'What is the smallest unit of matter?', 
 'single_choice', 
 '[{"text": "Molecule", "is_correct": false}, {"text": "Atom", "is_correct": true}, {"text": "Electron", "is_correct": false}, {"text": "Proton", "is_correct": false}]'::jsonb,
 '20', 4),

-- Question 6
((SELECT id FROM quizzes WHERE slug = 'science-technology-quiz'), 
 'Which technologies are used for artificial intelligence?', 
 'multiple_choice', 
 '[{"text": "Machine Learning", "is_correct": true}, {"text": "Neural Networks", "is_correct": true}, {"text": "Blockchain", "is_correct": false}, {"text": "Quantum Computing", "is_correct": false}]'::jsonb,
 '45', 5),

-- Question 7
((SELECT id FROM quizzes WHERE slug = 'science-technology-quiz'), 
 'What does DNA stand for?', 
 'text_input', 
 '[{"text": "Deoxyribonucleic Acid", "is_correct": true}, {"text": "deoxyribonucleic acid", "is_correct": true}]'::jsonb,
 '20', 6),

-- Question 8
((SELECT id FROM quizzes WHERE slug = 'science-technology-quiz'), 
 'Which company developed the iPhone?', 
 'single_choice', 
 '[{"text": "Samsung", "is_correct": false}, {"text": "Google", "is_correct": false}, {"text": "Apple", "is_correct": true}, {"text": "Microsoft", "is_correct": false}]'::jsonb,
 '20', 7),

-- Question 9
((SELECT id FROM quizzes WHERE slug = 'science-technology-quiz'), 
 'What is the powerhouse of the cell?', 
 'single_choice', 
 '[{"text": "Nucleus", "is_correct": false}, {"text": "Mitochondria", "is_correct": true}, {"text": "Ribosome", "is_correct": false}, {"text": "Endoplasmic Reticulum", "is_correct": false}]'::jsonb,
 '20', 8),

-- Question 10
((SELECT id FROM quizzes WHERE slug = 'science-technology-quiz'), 
 'Which of these are cloud computing platforms?', 
 'multiple_choice', 
 '[{"text": "AWS", "is_correct": true}, {"text": "Azure", "is_correct": true}, {"text": "Google Cloud", "is_correct": true}, {"text": "Oracle", "is_correct": false}]'::jsonb,
 '45', 9);

-- Quiz 3: History & Geography Quiz (10 questions)
INSERT INTO questions (quiz_id, question, type, answers, time_limit, index) VALUES
-- Question 1
((SELECT id FROM quizzes WHERE slug = 'history-geography-quiz'), 
 'In which year did the Berlin Wall fall?', 
 'single_choice', 
 '[{"text": "1987", "is_correct": false}, {"text": "1989", "is_correct": true}, {"text": "1991", "is_correct": false}, {"text": "1993", "is_correct": false}]'::jsonb,
 '20', 0),

-- Question 2
((SELECT id FROM quizzes WHERE slug = 'history-geography-quiz'), 
 'Which countries border France?', 
 'multiple_choice', 
 '[{"text": "Spain", "is_correct": true}, {"text": "Germany", "is_correct": true}, {"text": "Italy", "is_correct": true}, {"text": "Portugal", "is_correct": false}]'::jsonb,
 '45', 1),

-- Question 3
((SELECT id FROM quizzes WHERE slug = 'history-geography-quiz'), 
 'What is the capital of Canada?', 
 'text_input', 
 '[{"text": "Ottawa", "is_correct": true}, {"text": "ottawa", "is_correct": true}]'::jsonb,
 '20', 2),

-- Question 4
((SELECT id FROM quizzes WHERE slug = 'history-geography-quiz'), 
 'Who was the first person to walk on the moon?', 
 'single_choice', 
 '[{"text": "Buzz Aldrin", "is_correct": false}, {"text": "Neil Armstrong", "is_correct": true}, {"text": "John Glenn", "is_correct": false}, {"text": "Alan Shepard", "is_correct": false}]'::jsonb,
 '20', 3),

-- Question 5
((SELECT id FROM quizzes WHERE slug = 'history-geography-quiz'), 
 'Which river is the longest in the world?', 
 'single_choice', 
 '[{"text": "Amazon River", "is_correct": false}, {"text": "Nile River", "is_correct": true}, {"text": "Mississippi River", "is_correct": false}, {"text": "Yangtze River", "is_correct": false}]'::jsonb,
 '20', 4),

-- Question 6
((SELECT id FROM quizzes WHERE slug = 'history-geography-quiz'), 
 'Which ancient civilizations built pyramids?', 
 'multiple_choice', 
 '[{"text": "Egyptians", "is_correct": true}, {"text": "Mayans", "is_correct": true}, {"text": "Greeks", "is_correct": false}, {"text": "Romans", "is_correct": false}]'::jsonb,
 '45', 5),

-- Question 7
((SELECT id FROM quizzes WHERE slug = 'history-geography-quiz'), 
 'What is the largest country by land area?', 
 'text_input', 
 '[{"text": "Russia", "is_correct": true}, {"text": "russia", "is_correct": true}]'::jsonb,
 '20', 6),

-- Question 8
((SELECT id FROM quizzes WHERE slug = 'history-geography-quiz'), 
 'Which empire was ruled by Julius Caesar?', 
 'single_choice', 
 '[{"text": "Greek Empire", "is_correct": false}, {"text": "Roman Empire", "is_correct": true}, {"text": "Byzantine Empire", "is_correct": false}, {"text": "Ottoman Empire", "is_correct": false}]'::jsonb,
 '20', 7),

-- Question 9
((SELECT id FROM quizzes WHERE slug = 'history-geography-quiz'), 
 'Which of these cities are national capitals?', 
 'multiple_choice', 
 '[{"text": "Tokyo", "is_correct": true}, {"text": "Berlin", "is_correct": true}, {"text": "Sydney", "is_correct": false}, {"text": "New York", "is_correct": false}]'::jsonb,
 '45', 8),

-- Question 10
((SELECT id FROM quizzes WHERE slug = 'history-geography-quiz'), 
 'In which year did the Titanic sink?', 
 'single_choice', 
 '[{"text": "1910", "is_correct": false}, {"text": "1912", "is_correct": true}, {"text": "1914", "is_correct": false}, {"text": "1916", "is_correct": false}]'::jsonb,
 '20', 9);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Delete questions first (due to foreign key constraints)
DELETE FROM questions WHERE quiz_id IN (
    SELECT id FROM quizzes WHERE slug IN ('general-knowledge-quiz', 'science-technology-quiz', 'history-geography-quiz')
);

-- Delete quizzes
DELETE FROM quizzes WHERE slug IN ('general-knowledge-quiz', 'science-technology-quiz', 'history-geography-quiz');

-- +goose StatementEnd
