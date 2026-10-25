## About the Revision of October 25, 2026

We revised this Policy on October 25, 2026, and it took effect on the same day.
Users do not need to take any action.
If you do not agree to how we handle information after the revision, you can stop using the Service and leave it from the [account deletion page](/settings/account/deletion/new).

The points we clarified or changed in this revision are as follows:

- We combined the information we handle and the utilization purposes into one table that reflects how we actually handle them
- We set out how we handle information about people involved in the production of anime works, and where to request that their information be delisted
- We explained the information we provide through the Web API, and removed the provision on third-party provision by the opt-out method
- We removed the provision of data to third parties and use for marketing from the utilization purposes
- We specified the external services we use and the information transmitted to outside parties for advertising and other purposes
- We set out our security control action and how we handle logs
- We set out how to make a demand for disclosure, etc., how we verify your identity, and the fees
- We rewrote the information we delete and the information that remains for a certain period when you leave the Service, to reflect how we actually handle it
- We set out that we provide our address upon request
- We changed how we notify you when we change this Policy

The previous version is available in Japanese only: [Version effective February 16, 2019](/privacy/archive/20190216).

---

The operator (the "Operator") of the anime tracking service "Annict" (the "Service") sets out this Privacy Policy (this "Policy") on the personal information handled by the Service as follows.

This English version is a translation of the Japanese original.
If there is any discrepancy in interpretation between the Japanese version and the English version, the Japanese version prevails.
This Policy is governed by the laws of Japan, including the Act on the Protection of Personal Information.

## 1. Information We Handle

The Operator handles the following information about users of the Service:

- Account information: email address, username, and password (we store only its hashed value)
- Profile: name, bio, URL, profile image, and background image
- Posts and activities: records, reviews, watching statuses, library, likes, comments, follows, favorites, collections, forum posts, and edits on Annict DB
- Settings: display language, time zone, notification settings, and so on
- Supporter information: Stripe customer ID and subscription status
  - For supporters who subscribed through Gumroad in the past: Gumroad subscriber ID, email address, and subscription status
  - Users enter payment information such as card numbers directly on the screen of the payment provider (Stripe), and the Operator does not obtain it
- Inquiry information: the email address and content of inquiries
- Access information: IP address, URL of pages viewed, date and time of access, browser information (User-Agent and so on), and cookies

The Service uses cookies to keep you signed in and to distinguish devices when we release features gradually.

In addition, the Operator handles information about people involved in the production of anime works ("3. Handling of Information about People").

## 2. Utilization Purposes

The Operator utilizes the information it handles for the following purposes:

| Utilization purpose | Information utilized |
| --- | --- |
| Providing the Service (account registration and authentication, saving and displaying posts, and notifications) | Account information, profile, posts and activities, and settings |
| Accepting supporter subscriptions, providing benefits, and handling cancellations | Account information and supporter information |
| Responding to inquiries and sending important notices such as changes to the terms or this Policy | Account information and inquiry information |
| Dealing with unauthorized access and acts that violate the terms, and investigating and preventing the recurrence of failures | Account information, posts and activities, and access information |
| Creating statistics that do not identify individuals, used to improve the Service | Posts and activities, and access information |
| Displaying advertisements on the Service | Access information |

The information transmitted to advertising providers to display advertisements is described in "7. External Transmission".

## 3. Handling of Information about People

The Service publishes information about people involved in the production of anime works (the "People") as the cast and staff of the works.

- Information handled: name (including reading, English name, and aliases), gender, date of birth, blood type, height, place of birth, links to the official website, Wikipedia, and X account, and the works, roles, and characters the person was involved in
- Sources: publicly available information such as the official websites and official announcements of works, and profiles by the person or their agency
  - The Operator and editors of Annict DB (users of the Service) register and edit the information
- Utilization purposes: to publish the information on the Service as the cast and staff of works, and to provide it through the Web API
  - To allow users to search for People and add them to their favorites
- Scope of publication: anyone can view the information on the Service, and we also provide it through the Web API ("4. Provision through the Web API")

The person or their agent may request that the information be delisted or corrected.
Please contact us at the email address in "11. Contact".
How we verify identity is described in "10. Demands for Disclosure, etc.".

Once delisted, we no longer display the person's page or the person's rows in the cast and staff sections of works, and we no longer provide the information through the Web API.
To prevent the same person from being registered again, we keep the record of the delisted person instead of deleting it.
The Operator cannot delete information that remains in search engine caches or in other services that obtained it through the Web API.

## 4. Provision through the Web API

The Service also provides information published on the Service to other applications and services through the Web API (REST API and GraphQL API).

- Information provided: information that anyone can view on the Service, such as users' profiles, records, reviews, watching statuses, libraries, and follows, and information about People
- There is no setting for each user to stop the provision through the Web API
- Applications that a user has authorized can also obtain that user's email address
  - You can revoke the authorization at any time from the [connected applications settings](/settings/apps)

## 5. Provision to a Third Party

The Operator does not provide personal data to a third party without obtaining the principal's consent in advance, except in the following cases:

- Cases based on laws and regulations
- Cases in which there is a need to protect a human life, body, or fortune, and when it is difficult to obtain the principal's consent
- Cases in which there is a need to cooperate with a central government organization or the like in performing affairs prescribed by laws and regulations, and when there is a possibility that obtaining the principal's consent would interfere with the performance of those affairs
- Other cases permitted under the Act on the Protection of Personal Information
- Cases in which we entrust the handling of personal data within the scope necessary to achieve the utilization purposes ("6. External Services We Use")
- Cases in which personal data is provided accompanying business succession caused by a merger or other reason

## 6. External Services We Use

The Operator uses the following external services to operate the Service, and entrusts the handling of personal data to them within the necessary scope.
For services whose storage country cannot be identified, we describe the reason and reference information.

- XServer VPS (server)
  - Purpose: servers and databases of the Service
  - Information handled: all information handled by the Service
  - Country where the provider is located: Japan
  - Country where the information is stored: Japan
- Cloudflare (traffic relay and storage)
  - Purpose: relaying traffic, bot protection, and storing images and database backups
  - Information handled: traffic content, IP addresses, images uploaded by users, and database backups
  - Country where the provider is located: United States
  - Country where the information is stored: Cannot be identified
    - This is because Cloudflare discloses the storage location only down to the region (Asia-Pacific)
- Sentry (error monitoring)
  - Purpose: detecting and investigating errors
  - Information handled: user ID, username, IP address, URL of the page where the error occurred, and browser information
  - Country where the provider is located: United States
  - Country where the information is stored: United States
- Resend (email delivery)
  - Purpose: sending emails such as verification codes, password resets, and notifications
  - Information handled: email addresses, and email subjects and bodies
  - Country where the provider is located: United States
  - Country where the information is stored: Japan
- Stripe (payments)
  - Purpose: supporter payments
  - Information handled: user ID, Stripe customer ID, and display language
  - Country where the provider is located: United States (our contracting party in Japan is Stripe Japan, Inc.)
  - Country where the information is stored: Cannot be identified
    - This is because Stripe does not disclose the country where it stores information
    - For details, see [Stripe's Privacy Policy](https://stripe.com/privacy)
- Google Workspace (receiving email)
  - Purpose: receiving and storing inquiry emails
  - Information handled: the email address and content of inquiries
  - Country where the provider is located: United States
  - Country where the information is stored: Cannot be identified
    - This is because Google does not disclose the country where it stores information
    - For details, see [Google's Privacy Policy](https://policies.google.com/privacy?hl=en)
- Discord (notifications to the Operator)
  - Purpose: notifying the Operator of new forum posts
  - Information handled: the poster's name, and the title and URL of the post
  - Country where the provider is located: United States
  - Country where the information is stored: United States (it may also be stored on servers in other countries)

We will provide, upon request, information on the action taken by providers in foreign countries for the protection of personal information.
Please contact us at the email address in "11. Contact".

## 7. External Transmission

When you open a page of the Service, your browser transmits information to the following providers:

- Google AdSense (Google LLC)
  - Purpose: displaying advertisements
  - Information transmitted: IP address, browser information (User-Agent and so on), URL of the page viewed, and Google's advertising cookies
  - Pages that transmit the information: most pages of the Service
  - The Operator has disabled personalized advertising
  - For how Google uses the information, see [Google's explanation](https://policies.google.com/technologies/ads?hl=en)
- Google Fonts (Google LLC)
  - Purpose: displaying fonts
  - Information transmitted: IP address, browser information (User-Agent and so on), and the domain of the site being viewed
  - Pages that transmit the information: most pages of the Service
- Cloudflare Turnstile (Cloudflare, Inc.)
  - Purpose: preventing unauthorized operations by bots
  - Information transmitted: IP address, browser information (User-Agent and so on), and information on browser behavior used to distinguish bots
  - Pages that transmit the information: the sign-in, sign-up, and password reset pages

## 8. Handling of Logs

The Service records the IP address, the page viewed, and the date and time in logs for each access.
Values such as email addresses, passwords, and tokens are masked before being recorded.
IP addresses are used to investigate unauthorized access and access concentrated in a short period.
Logs are stored for one year and then deleted.

## 9. Security Control Action

The Operator takes the following action to prevent leakage, etc. of personal data:

- Organizational action
  - The Operator also serves as the person responsible for the handling of personal information
  - The permissions of Annict DB editors are limited to editing work data, and there are no screens that display users' personal information to them
  - Work involving data in the production environment is carried out with procedures and records
  - We publish a [vulnerability reporting contact](https://github.com/annict/annict/security/policy)
- Human action
  - Only the Operator can operate the production environment (servers, databases, storage, and the admin screens of external services)
- Physical action
  - We do not own our own facilities and use providers' servers and cloud services
  - The Operator's work devices have disk encryption and screen lock enabled
- Technical action
  - Communications are encrypted with HTTPS, and connections over HTTP are redirected to HTTPS
    - Communications from Cloudflare to our servers are also encrypted
  - Passwords and verification codes are stored only as hashed values
    - IDs used to keep you signed in and tokens used to reset passwords are also stored as hashed values
  - Cookies that keep you signed in are sent only over HTTPS and cannot be read by JavaScript
  - Sign-in, sign-up, and password reset have attempt limits and bot protection
  - Connections to databases are limited to those from within the Operator's VPN
  - Database backups are taken daily and stored for 30 days
  - Logs and error monitoring mask values such as email addresses, passwords, and tokens
  - We receive vulnerability information on the software we use and update it
  - Secrets such as passwords are not included in source code and are managed in a password management service
  - Copies of the production database used for development are used only after replacing email addresses, IP addresses, passwords, and so on
- Understanding the external environment
  - We use services of providers in foreign countries
    The countries where the providers are located and where the information is stored are as described in "6. External Services We Use"
  - We take the above action after understanding the personal information protection systems of the United States through the [Personal Information Protection Commission's survey](https://www.ppc.go.jp/enforcement/infoprovision/laws/offshore_report_america/) (in Japanese)

## 10. Demands for Disclosure, etc.

The principal or their agent may demand the following regarding the retained personal data held by the Operator:

- Notification of the utilization purpose
- Disclosure (including disclosure of the records of provision to a third party)
- Correction, addition, or deletion when the content is not factual
- Utilization cease or deletion
- Cease of the provision to a third party

Please send your demand to the email address in "11. Contact".
There is no prescribed form.
Please state the type of demand and the information it concerns.

- Identity verification
  - For users, the Operator sends a confirmation email to the registered email address and verifies your identity by your reply
  - For People, we verify identity through a contact channel that the person uses publicly, such as their official website, official social media account, or agency
  - For agents, we verify the relationship with the principal and the fact of the delegation
  - We do not ask for images of identification documents or the like
- Fees: free of charge
- Method of response: in principle, we respond by email
  - Information to be disclosed is provided as electronic files

If we are not obligated by laws and regulations to respond to the demand, or if we cannot verify identity, we will notify you of that and the reason.

## 11. Contact

Please send opinions, questions, complaints, and demands regarding this Policy and the handling of personal information to the following contact:

- Operator: Koji Shimba
- Email address: hello@annict.com

The Operator's address will be provided without delay upon request.

## 12. Handling When You Leave the Service

Users can leave the Service from the [account deletion page](/settings/account/deletion/new).
When you leave, we delete your account information, profile, posts and activities, and settings.
Deletion is carried out in sequence after the leaving procedure, so it may take some time to complete.

The following information remains after you leave:

- Your username as text written in other users' posts
- The bodies of forum comments you wrote in other users' threads (the link to your account is removed)
- The content of work data you edited on Annict DB and the work images you uploaded (the link to your account is removed)
- Records of sign-up verification codes (including email addresses)
- Records of payment notifications received from Stripe (which may include customer IDs and so on)

There is no retention period for records of verification codes and payment notifications.

The following information remains for a certain period after you leave and is deleted after the period ends:

- Database backups: 30 days
- Logs: 1 year
- Error monitoring (Sentry) records: 30 days
- Email delivery (Resend) records: 30 days
- Records of sign-in sessions on other devices: 30 days after the last access
- Profile images copied to the development environment: until the next time the copy is refreshed

When a supporter who subscribed through Stripe leaves, we cancel the subscription at that time.
Payment records on Stripe are stored in accordance with Stripe's rules.

Subscriptions of supporters who subscribed through Gumroad are not canceled when they leave.
Please cancel them on Gumroad.

## 13. Changes to This Policy

The Operator may change this Policy in response to amendments to laws and regulations or changes to the Service.
When we change it, we will announce it on the Service or in the Help, and also on the Service's official social media accounts.
When a change requires the principal's consent under laws and regulations, we will obtain consent.

## Supplementary Provisions

- Established: March 10, 2014
- Revised: February 16, 2019, and October 25, 2026
- Effective: October 25, 2026

The revision history is as follows (available in Japanese only):

- [Version effective March 10, 2014](/privacy/archive/20140310)
- [Version effective February 16, 2019](/privacy/archive/20190216)
