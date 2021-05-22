# frozen_string_literal: true

USERNAME_FORMAT = /[A-Za-z0-9_]+/

get "dummy_image", to: "application#dummy_image" if Rails.env.test?

devise_for :users,
  controllers: {omniauth_callbacks: "v6/callbacks"},
  skip: %i[passwords registrations sessions]

use_doorkeeper do
  controllers applications: "oauth/applications"
  skip_controllers :authorized_applications
end

resources :comments, only: %i[edit update destroy]
resources :mute_users, only: [:destroy]

resources :channels, only: [:index]

resources :episodes, only: [] do
  resources :records, only: [], controller: :episode_records do
    post :switch, on: :collection
  end
end

resources :users, only: [] do
  collection do
    delete :destroy
  end
end

scope "@:username", username: USERNAME_FORMAT do
  get :ics, to: "ics#show", as: :user_ics

  resources :tags, only: %i[show], controller: :user_work_tags, as: :user_work_tag
  resources :reviews, only: %i[show]

  resources :records, only: %i[] do
    resources :comments, only: %i[create]
  end
end

resources :works, only: %i[index] do
  resources :episodes, only: [] do
    resources :checkins, only: %i[show]
  end
end

# 新リダイレクト用URL
get "r/:provider/:url_hash",
  to: "episode_records#redirect",
  provider: /fb|tw/,
  url_hash: /[0-9a-zA-Z_-]{10}/

devise_scope :user do
  match "/legacy/sign_in", via: :get, as: :legacy_sign_in, to: "v6/legacy/sessions#new"
  match "/legacy/sign_in", via: :post, as: :user_session, to: "v6/legacy/sessions#create"
  match "/sign_out", via: :delete, as: :sign_out, to: "devise/sessions#destroy"
  match "/user_email", via: :patch, as: :user_email, to: "v4/user_emails#update"
  match "/user_email/callback", via: :get, as: :user_email_callback, to: "v4/user_email_callbacks#show"
end

match "/@:username", via: :get, as: :profile, to: "v6/users#show", username: USERNAME_FORMAT
match "/@:username/:status_kind", via: :get, as: :library, to: "v4/libraries#show", username: USERNAME_FORMAT, status_kind: /wanna_watch|watching|watched|on_hold|stop_watching/
match "/@:username/favorite_characters", via: :get, as: :favorite_character_list, to: "v3/favorite_characters#index", username: USERNAME_FORMAT
match "/@:username/favorite_organizations", via: :get, as: :favorite_organization_list, to: "v3/favorite_organizations#index", username: USERNAME_FORMAT
match "/@:username/favorite_people", via: :get, as: :favorite_person_list, to: "v3/favorite_people#index", username: USERNAME_FORMAT
match "/@:username/followers", via: :get, as: :followers_user, to: "v3/users#followers", username: USERNAME_FORMAT
match "/@:username/following", via: :get, as: :following_user, to: "v3/users#following", username: USERNAME_FORMAT
match "/@:username/records", via: :get, as: :record_list, to: "v4/records#index", username: USERNAME_FORMAT
match "/@:username/records/:record_id", via: :delete, to: "v4/records#destroy", username: USERNAME_FORMAT
match "/@:username/records/:record_id", via: :get, to: "v4/records#show", username: USERNAME_FORMAT
match "/@:username/records/:record_id", via: :patch, as: :record, to: "records#update", username: USERNAME_FORMAT
match "/@:username/records/:record_id", via: :patch, to: "v4/records#update", username: USERNAME_FORMAT
match "/characters/:character_id", via: :get, as: :character, to: "v3/characters#show"
match "/characters/:character_id/fans", via: :get, as: :character_fan_list, to: "v3/character_fans#index"
match "/episode_records", via: :patch, as: :episode_record_mutation, to: "v4/episode_records#update"
match "/episodes/:episode_id/records", via: :post, as: :episode_record_list, to: "episode_records#create"
match "/faq", via: :get, as: :faq, to: "v6/faqs#show"
match "/fragment/@:username/records/:record_id/edit", via: :get, as: :fragment_edit_record, to: "v6/fragment/records#edit", username: USERNAME_FORMAT
match "/fragment/activity_groups/:activity_group_id/items", via: :get, as: :fragment_activity_item_list, to: "v6/fragment/activity_items#index"
match "/fragment/episodes/:episode_id/records", via: :get, as: :fragment_episode_record_list, to: "v6/fragment/episode_records#index"
match "/fragment/receive_channel_buttons", via: :get, as: :fragment_receive_channel_button_list, to: "v6/fragment/receive_channel_buttons#index"
match "/fragment/trackable_anime/:anime_id", via: :get, as: :fragment_trackable_anime, to: "v6/fragment/trackable_anime#show"
match "/fragment/trackable_episodes", via: :get, as: :fragment_trackable_episode_list, to: "v6/fragment/trackable_episodes#index"
match "/fragment/trackable_episodes/:episode_id", via: :get, as: :fragment_trackable_episode, to: "v6/fragment/trackable_episodes#show"
match "/friends", via: :get, as: :friend_list, to: "v3/friends#index"
match "/legal", via: :get, as: :legal, to: "v6/pages#legal"
match "/notifications", via: :get, as: :notification_list, to: "v3/notifications#index"
match "/organizations/:organization_id", via: :get, as: :organization, to: "v3/organizations#show"
match "/organizations/:organization_id/fans", via: :get, as: :organization_fan_list, to: "v3/organization_fans#index"
match "/people/:person_id", via: :get, as: :person, to: "v3/people#show"
match "/people/:person_id/fans", via: :get, as: :person_fan_list, to: "v3/person_fans#index"
match "/privacy", via: :get, as: :privacy, to: "v6/pages#privacy"
match "/registrations/new", via: :get, as: :new_registration, to: "v6/registrations#new"
match "/search", via: :get, as: :search, to: "v3/searches#show"
match "/settings", via: :get, as: :setting_list, to: "v3/settings#index"
match "/settings/account", via: :get, as: :settings_account, to: "v3/settings/accounts#show"
match "/settings/account", via: :patch, to: "v3/settings/accounts#update"
match "/settings/apps", via: :get, as: :settings_app_list, to: "v3/settings/apps#index"
match "/settings/apps/:app_id/revoke", via: :patch, as: :settings_revoke_app, to: "v3/settings/apps#revoke"
match "/settings/email_notification", via: :get, as: :settings_email_notification, to: "v3/settings/email_notifications#show"
match "/settings/email_notification", via: :patch, to: "v3/settings/email_notifications#update"
match "/settings/email_notification/unsubscribe", via: :get, as: :settings_unsubscribe_email_notification, to: "v3/settings/email_notifications#unsubscribe"
match "/settings/muted_users", via: :get, as: :settings_muted_user_list, to: "v3/settings/muted_users#index"
match "/settings/options", via: :get, as: :settings_option_list, to: "v3/settings/options#index"
match "/settings/options", via: :patch, to: "v3/settings/options#update"
match "/settings/password", via: :patch, as: :settings_password, to: "v3/settings/passwords#update"
match "/settings/profile", via: :get, as: :settings_profile, to: "v3/settings/profiles#show"
match "/settings/profile", via: :patch, to: "v3/settings/profiles#update"
match "/settings/providers", via: :get, as: :settings_provider_list, to: "v3/settings/providers#index"
match "/settings/providers/:provider_id", via: :delete, as: :settings_provider, to: "v3/settings/providers#destroy"
match "/settings/tokens", via: :post, as: :settings_token_list, to: "v3/settings/tokens#create"
match "/settings/tokens/:token_id", via: :delete, as: :settings_token, to: "v3/settings/tokens#destroy"
match "/settings/tokens/:token_id", via: :patch, to: "v3/settings/tokens#update"
match "/settings/tokens/:token_id/edit", via: :get, as: :settings_edit_token, to: "v3/settings/tokens#edit"
match "/settings/tokens/new", via: :get, as: :settings_new_token, to: "v3/settings/tokens#new"
match "/sign_in", via: :get, as: :new_user_session, to: "v6/sign_in#new" # for Devise
match "/sign_in", via: :get, as: :sign_in, to: "v6/sign_in#new"
match "/sign_in/callback", via: :get, as: :sign_in_callback, to: "v6/sign_in_callbacks#show"
match "/sign_up", via: :get, as: :sign_up, to: "v6/sign_up#new"
match "/supporters", via: :get, as: :supporters, to: "v3/supporters#show"
match "/terms", via: :get, as: :terms, to: "v6/pages#terms"
match "/track", via: :get, as: :track, to: "tracks#show"
match "/work_display_option", via: :get, as: :work_display_option, to: "v3/work_display_options#show"
match "/works/:anime_id", via: :get, as: :anime, to: "v4/works#show", anime_id: /[0-9]+/
match "/works/:anime_id/episodes", via: :get, as: :episode_list, to: "v4/episodes#index", anime_id: /[0-9]+/
match "/works/:anime_id/episodes/:episode_id", via: :get, as: :episode, to: "v6/episodes#show", anime_id: /[0-9]+/
match "/works/:anime_id/records", via: :get, as: :anime_record_list, to: "v4/anime_records#index", anime_id: /[0-9]+/
match "/works/:anime_id/records", via: :post, to: "v4/anime_records#create", anime_id: /[0-9]+/
match "/works/:slug", via: :get, as: :seasonal_anime_list, to: "v3/works#season", slug: /[0-9]{4}-(all|spring|summer|autumn|winter)/
match "/works/newest", via: :get, as: :newest_anime_list, to: "v3/works#newest"
match "/works/popular", via: :get, as: :popular_anime_list, to: "v3/works#popular"

root "v6/home#show",
  constraints: Annict::RoutingConstraints::Member.new
root "v6/welcome#show",
  constraints: Annict::RoutingConstraints::Guest.new,
  # Set :as option to avoid two routes with the same name
  as: nil
