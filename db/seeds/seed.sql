-- Seed stations
INSERT INTO stations (id, name, latitude, longitude) VALUES
    ('stn-seed-001', 'North Field A',         38.5767,  -93.2650),
    ('stn-seed-002', 'South Greenhouse',       37.0902,  -95.7129),
    ('stn-seed-003', 'East Orchard Row 3',     35.4676,  -97.5164),
    ('stn-seed-004', 'West Irrigation Zone',   36.1540,  -94.1323)
ON CONFLICT (id) DO NOTHING;

-- Seed readings for stn-seed-001 (North Field A)
INSERT INTO readings (station_id, type, value, recorded_at) VALUES
    ('stn-seed-001', 'temperature',   78.4, now() - interval '2 minutes'),
    ('stn-seed-001', 'temperature',   77.9, now() - interval '17 minutes'),
    ('stn-seed-001', 'temperature',   76.1, now() - interval '32 minutes'),
    ('stn-seed-001', 'humidity',      61.2, now() - interval '2 minutes'),
    ('stn-seed-001', 'humidity',      62.5, now() - interval '17 minutes'),
    ('stn-seed-001', 'humidity',      64.0, now() - interval '32 minutes'),
    ('stn-seed-001', 'soil_moisture', 43.7, now() - interval '2 minutes'),
    ('stn-seed-001', 'soil_moisture', 44.1, now() - interval '17 minutes'),
    ('stn-seed-001', 'soil_moisture', 44.8, now() - interval '32 minutes'),
    ('stn-seed-001', 'wind_direction', 225,  now() - interval '2 minutes'),
    ('stn-seed-001', 'wind_direction', 218,  now() - interval '17 minutes'),
    ('stn-seed-001', 'wind_direction', 230,  now() - interval '32 minutes');

-- Seed readings for stn-seed-002 (South Greenhouse)
INSERT INTO readings (station_id, type, value, recorded_at) VALUES
    ('stn-seed-002', 'temperature',   84.1, now() - interval '3 minutes'),
    ('stn-seed-002', 'temperature',   83.6, now() - interval '18 minutes'),
    ('stn-seed-002', 'temperature',   82.0, now() - interval '33 minutes'),
    ('stn-seed-002', 'humidity',      72.8, now() - interval '3 minutes'),
    ('stn-seed-002', 'humidity',      71.3, now() - interval '18 minutes'),
    ('stn-seed-002', 'humidity',      70.5, now() - interval '33 minutes'),
    ('stn-seed-002', 'soil_moisture', 58.2, now() - interval '3 minutes'),
    ('stn-seed-002', 'soil_moisture', 57.9, now() - interval '18 minutes'),
    ('stn-seed-002', 'soil_moisture', 57.4, now() - interval '33 minutes'),
    ('stn-seed-002', 'wind_direction',  45,  now() - interval '3 minutes'),
    ('stn-seed-002', 'wind_direction',  52,  now() - interval '18 minutes'),
    ('stn-seed-002', 'wind_direction',  40,  now() - interval '33 minutes');

-- Seed readings for stn-seed-003 (East Orchard Row 3)
INSERT INTO readings (station_id, type, value, recorded_at) VALUES
    ('stn-seed-003', 'temperature',   91.3, now() - interval '1 minute'),
    ('stn-seed-003', 'temperature',   90.7, now() - interval '16 minutes'),
    ('stn-seed-003', 'humidity',      38.4, now() - interval '1 minute'),
    ('stn-seed-003', 'humidity',      39.1, now() - interval '16 minutes'),
    ('stn-seed-003', 'soil_moisture', 29.6, now() - interval '1 minute'),
    ('stn-seed-003', 'soil_moisture', 30.2, now() - interval '16 minutes'),
    ('stn-seed-003', 'wind_direction', 135,  now() - interval '1 minute'),
    ('stn-seed-003', 'wind_direction', 128,  now() - interval '16 minutes');

-- Seed readings for stn-seed-004 (West Irrigation Zone)
INSERT INTO readings (station_id, type, value, recorded_at) VALUES
    ('stn-seed-004', 'temperature',   67.5, now() - interval '5 minutes'),
    ('stn-seed-004', 'temperature',   66.8, now() - interval '20 minutes'),
    ('stn-seed-004', 'humidity',      55.3, now() - interval '5 minutes'),
    ('stn-seed-004', 'humidity',      56.0, now() - interval '20 minutes'),
    ('stn-seed-004', 'soil_moisture', 66.9, now() - interval '5 minutes'),
    ('stn-seed-004', 'soil_moisture', 67.4, now() - interval '20 minutes'),
    ('stn-seed-004', 'wind_direction', 315,  now() - interval '5 minutes'),
    ('stn-seed-004', 'wind_direction', 310,  now() - interval '20 minutes');
