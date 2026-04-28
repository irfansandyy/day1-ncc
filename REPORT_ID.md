# Laporan Day 2 - Jenkins + SonarQube

## Ringkasan Pemenuhan

- Menyiapkan Jenkins sebagai tools automation: terpenuhi (Jenkins pipeline dan stack Docker).
- Menyiapkan SonarQube sebagai tools code quality analysis: terpenuhi (service SonarQube + scanner).
- Menghubungkan Jenkins dengan SonarQube dan membuat pipeline: terpenuhi (`withSonarQubeEnv` + `sonar-scanner`).

## Poin Opsional

- Jenkins Pipeline dibanding freestyle: terpenuhi (Jenkinsfile).
- Stage terstruktur build, test, analyze + Quality Gate: terpenuhi (Quality Gate abort pipeline jika gagal).
- Webhook trigger saat push: ditambahkan via `githubPush()` dan panduan webhook.
- Environment variable / credential management: digunakan (parameter `GIT_CREDENTIALS_ID`, kredensial SonarQube di Jenkins).
- Build badge/status: disediakan URL badge Jenkins (lihat bagian badge di README).
- Optimasi pipeline: ditambahkan parallel stage untuk backend build dan test.

## Deskripsi Pipeline

Pipeline menggunakan Jenkins declarative pipeline dengan tahapan berikut:

1. Checkout repository (branch dan URL dapat diatur lewat parameter).
2. Setup backend dependencies (Go modules).
3. Backend Build & Test dijalankan paralel untuk mempercepat waktu build.
4. Frontend install dependencies, lint, lalu build.
5. SonarQube analysis untuk backend dan frontend.
6. Quality Gate menahan/ menggagalkan pipeline jika standar kualitas tidak terpenuhi.

## Integrasi Jenkins dengan SonarQube

- Jenkins dikonfigurasi dengan server SonarQube bernama `SonarQube`.
- Jenkinsfile memanggil `withSonarQubeEnv("SonarQube")` untuk inject URL dan token.
- `sonar-scanner` menjalankan analisis dan mengirim hasil ke SonarQube.
- `waitForQualityGate abortPipeline: true` memvalidasi status Quality Gate sebelum pipeline dinyatakan sukses.

## Alur Pipeline (Flow)

Checkout -> Setup -> Backend Build & Test (parallel) -> Frontend Install -> Frontend Lint -> Frontend Build -> SonarQube Analysis -> Quality Gate -> Post Actions (archive artifacts).

## Webhook, Credentials, dan Badge

- Webhook GitHub: gunakan endpoint `http(s)://<jenkins-host>/jenkins/github-webhook/`.
- Credentials: gunakan `GIT_CREDENTIALS_ID` untuk repo private; token SonarQube disimpan di Jenkins credentials.
- Build badge: `http(s)://<jenkins-host>/jenkins/job/<job-name>/badge/icon`.

## Screenshot (Tempel di bawah ini)

1. Konfigurasi Jenkins (Global Tool + SonarQube Server):
	- [tempel screenshot]
2. Konfigurasi Job Pipeline Jenkins:
	- [tempel screenshot]
3. Konfigurasi SonarQube Project / Token:
	- [tempel screenshot]
4. Hasil analisis di SonarQube (Overview + Quality Gate):
	- [tempel screenshot]

## Kendala

- Belum ada kendala yang tercatat.
