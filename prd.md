# Product Requirements Document (PRD)

**Product Name:** SI-TEPAT (Sistem Integrasi e-KTP Tepat Sasaran / e-KTP Targeted Integration System)
**Version:** 1.1 (MVP - Vercel Serverless Architecture)

---

## 1. Product Vision & Objectives

**Vision:** Modernize government bureaucracy and eliminate the need for physical KTP (National ID Card) photocopies in public service processes and subsidy distribution validation.

**Key Objectives:**

- Ensure targeted subsidy distribution (LPG 3kg, fuel) based on real-time population data.
- Prevent fraud, data manipulation, or duplicate claims through instant field validation.
- Provide an efficient, cost-effective, and auto-scaling infrastructure without traditional server management.

---

## 2. Target Users (User Personas)

| Persona | Description | Needs |
|---|---|---|
| **Field Officer (Merchant)** | Gas station attendants or LPG distribution agents | A fast, responsive, and easy-to-use mobile application for scanning e-KTP |
| **Government / Sub-district Admin** | Officials responsible for registering citizens' e-KTP and monitoring quotas | A comprehensive web dashboard accessible from anywhere |
| **Citizen (Passive End-User)** | e-KTP holders | Simply tap their e-KTP on the field officer's device |

---

## 3. Functional Requirements

### A. Mobile App (Field Officer)

- **Authentication:** Officers must be able to log in to the application using secure credentials.
- **NFC Scanner:** The application must be able to read the random UID (Unique Identifier) from the e-KTP chip using the NFC sensor on an Android smartphone.
- **Commodity Selection & Transaction:** Officers can select the subsidy type (e.g., LPG 3KG) and see an instant success/failure response on screen.

### B. Backend API Service (Golang - Serverless)

- **API Authentication:** Validate JWT Tokens from field officer application requests.
- **Data Validation:** Process e-KTP UID requests and validate them against the population database in Neon DB.
- **Quota Management (Concurrency Safe):** Execute queries that are safe from race conditions when deducting subsidy quotas in the database.
- **Serverless Format:** The API must be structured to be compatible for running as a single function (Serverless Functions) when deployed to Vercel.

### C. Admin Dashboard (Web)

- **Citizen Registration:** Interface for mapping a new citizen's e-KTP UID to their NIK (National ID Number).
- **Eligibility Management:** Feature to change a citizen's subsidy eligibility status (Active/Inactive).
- **Transaction Monitoring:** Real-time subsidy claim history monitoring table.

---

## 4. Non-Functional Requirements

- **Scalability & Availability:** The backend system and web frontend must be deployed on Vercel to ensure High Availability and auto-scaling capability from zero to thousands of requests without manual intervention.
- **Performance:** Golang API functions (Cold Start & Warm Start) must be optimized to deliver responses in milliseconds.
- **Security:** All data communication uses HTTPS. Neon DB Connection Strings and JWT Secret Keys must be stored as Environment Variables in the Vercel dashboard.

---

## 5. Technology Specifications (Tech Stack)

> [!IMPORTANT]
> The infrastructure adopts a **Serverless Ecosystem** approach.

| Component | Technology |
|---|---|
| **Mobile Client** | Flutter SDK (Dart) integrated with the `flutter_nfc_kit` library |
| **Backend API** | Golang (Implemented as Vercel Go Serverless Functions) |
| **Database Engine** | Neon DB (Serverless PostgreSQL) |
| **Database ORM/Query** | GORM or Golang standard library `database/sql` |
| **Admin Frontend (Web)** | Next.js (React) + Tailwind CSS |
| **Deployment Platform (CI/CD)** | Vercel (Directly connected to the GitHub repository for automatic deployment on every code update) |

---

## 6. System Constraints (Out of Scope)

> [!WARNING]
> The following items are **not** included in the scope of this MVP.

- No containerization (such as Docker) is used during the development and production phases.
- Local development is performed directly using native runtime commands (`go run` and `npm run dev`).
- No integration with the government's Dukcapil biometric API; only UID Chip validation is used.
