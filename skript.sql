-- Resetovanje ID sekvence za tablicu "Users"
TRUNCATE TABLE public.users RESTART IDENTITY CASCADE;

INSERT INTO public.users(
	username, password, email, role, name, surname, biography, moto, photo_url, is_blocked)
	VALUES ('admin', '$2a$10$2vlv2lL8AaBS6gGEXCDwhuCRPZLWo4c0AdZr8FHybKwe5UHlredEe', 'admin@gmail.com', 'Admin', 'Ime', 'Prezime', 'Biografija', 'Moto', 'Slika', false);

INSERT INTO public.users(
	username, password, email, role, name, surname, biography, moto, photo_url, is_blocked)
	VALUES ('turista', '$2a$10$2vlv2lL8AaBS6gGEXCDwhuCRPZLWo4c0AdZr8FHybKwe5UHlredEe', 'turista@gmail.com', 'Tourist', 'Ime', 'Prezime', 'Biografija', 'Moto', 'Slika', false);

INSERT INTO public.users(
	username, password, email, role, name, surname, biography, moto, photo_url, is_blocked)
	VALUES ('vodic', '$2a$10$2vlv2lL8AaBS6gGEXCDwhuCRPZLWo4c0AdZr8FHybKwe5UHlredEe', 'vodic@gmail.com', 'Guide', 'Ime', 'Prezime', 'Biografija', 'Moto', 'Slika', false);