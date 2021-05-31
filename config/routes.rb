# frozen_string_literal: true

ID_FORMAT = /[0-9]+/
USERNAME_FORMAT = /[A-Za-z0-9_]+/

Rails.application.routes.draw do
  draw :api
  draw :internal_api
  draw :local_api

  devise_for :users,
    controllers: {omniauth_callbacks: "v6/callbacks"},
    skip: %i[passwords registrations sessions]

  scope module: :v3 do
    use_doorkeeper do
      controllers(
        applications: "oauth/applications",
        authorizations: "oauth/authorizations",
        token_info: "oauth/token_info",
        tokens: "oauth/tokens"
      )
      skip_controllers :authorized_applications
    end
  end

  devise_scope :user do
    # standard:disable Layout/ExtraSpacing, Layout/LineLength
    match "/legacy/sign_in",      via: :get,    as: :legacy_sign_in,      to: "v6/legacy/sessions#new"
    match "/legacy/sign_in",      via: :post,   as: :user_session,        to: "v6/legacy/sessions#create"
    match "/sign_out",            via: :delete, as: :sign_out,            to: "devise/sessions#destroy"
    match "/user_email",          via: :patch,  as: :user_email,          to: "v4/user_emails#update"
    match "/user_email/callback", via: :get,    as: :user_email_callback, to: "v4/user_email_callbacks#show"
    # standard:enable Layout/ExtraSpacing, Layout/LineLength
  end

  # standard:disable Layout/ExtraSpacing, Layout/LineLength
  match "/@:username",                                        via: :get,    as: :profile,                                 to: "v6/users#show",                   username: USERNAME_FORMAT
  match "/@:username/:status_kind",                           via: :get,    as: :library,                                 to: "v4/libraries#show",               username: USERNAME_FORMAT, status_kind: /wanna_watch|watching|watched|on_hold|stop_watching/
  match "/@:username/favorite_characters",                    via: :get,    as: :favorite_character_list,                 to: "v3/favorite_characters#index",    username: USERNAME_FORMAT
  match "/@:username/favorite_organizations",                 via: :get,    as: :favorite_organization_list,              to: "v3/favorite_organizations#index", username: USERNAME_FORMAT
  match "/@:username/favorite_people",                        via: :get,    as: :favorite_person_list,                    to: "v3/favorite_people#index",        username: USERNAME_FORMAT
  match "/@:username/followers",                              via: :get,    as: :followers_user,                          to: "v3/users#followers",              username: USERNAME_FORMAT
  match "/@:username/following",                              via: :get,    as: :following_user,                          to: "v3/users#following",              username: USERNAME_FORMAT
  match "/@:username/ics",                                    via: :get,    as: :user_ics,                                to: "v3/ics#show",                     username: USERNAME_FORMAT
  match "/@:username/records",                                via: :get,    as: :record_list,                             to: "v4/records#index",                username: USERNAME_FORMAT
  match "/@:username/records/:record_id",                     via: :delete,                                               to: "v4/records#destroy",              username: USERNAME_FORMAT
  match "/@:username/records/:record_id",                     via: :get,                                                  to: "v4/records#show",                 username: USERNAME_FORMAT
  match "/@:username/records/:record_id",                     via: :patch,                                                to: "v4/records#update",               username: USERNAME_FORMAT
  match "/@:username/records/:record_id",                     via: :patch,  as: :record,                                  to: "records#update",                  username: USERNAME_FORMAT
  match "/api/internal/channels/:channel_id/reception",       via: :delete, as: :internal_api_channel_reception,          to: "api/internal/receptions#destroy"
  match "/api/internal/channels/:channel_id/reception",       via: :post,                                                 to: "api/internal/receptions#create"
  match "/api/internal/characters",                           via: :get,    as: :internal_api_character_list,             to: "api/internal/characters#index"
  match "/api/internal/follow",                               via: :delete, as: :internal_api_follow,                     to: "api/internal/follows#destroy"
  match "/api/internal/follow",                               via: :post,                                                 to: "api/internal/follows#create"
  match "/api/internal/following",                            via: :get,    as: :internal_api_following_list,             to: "api/internal/following#index"
  match "/api/internal/mute_user",                            via: :delete, as: :internal_api_mute_user,                  to: "api/internal/mute_users#destroy"
  match "/api/internal/mute_user",                            via: :post,                                                 to: "api/internal/mute_users#create"
  match "/api/internal/muted_users",                          via: :get,    as: :internal_api_muted_user_list,            to: "api/internal/muted_users#index"
  match "/api/internal/organizations",                        via: :get,    as: :internal_api_organization_list,          to: "api/internal/organizations#index"
  match "/api/internal/people",                               via: :get,    as: :internal_api_person_list,                to: "api/internal/people#index"
  match "/api/internal/series_list",                          via: :get,    as: :internal_api_series_list,                to: "api/internal/series_list#index"
  match "/api/internal/works",                                via: :get,    as: :internal_api_work_list,                  to: "api/internal/works#index"
  match "/channels",                                          via: :get,    as: :channel_list,                            to: "v6/channels#index"
  match "/characters/:character_id",                          via: :get,    as: :character,                               to: "v3/characters#show"
  match "/characters/:character_id/fans",                     via: :get,    as: :character_fan_list,                      to: "v3/character_fans#index"
  match "/db",                                                via: :get,    as: :db_root,                                 to: "v4/db/home#show"
  match "/db/activities",                                     via: :get,    as: :db_activity_list,                        to: "v4/db/activities#index"
  match "/db/casts/:id",                                      via: :delete, as: :db_cast_detail,                          to: "v4/db/casts#destroy"
  match "/db/casts/:id",                                      via: :patch,                                                to: "v4/db/casts#update"
  match "/db/casts/:id/edit",                                 via: :get,    as: :db_edit_cast,                            to: "v4/db/casts#edit"
  match "/db/casts/:id/publishing",                           via: :delete, as: :db_cast_publishing,                      to: "v4/db/cast_publishings#destroy"
  match "/db/casts/:id/publishing",                           via: :post,                                                 to: "v4/db/cast_publishings#create"
  match "/db/channel_groups",                                 via: :get,    as: :db_channel_group_list,                   to: "v4/db/channel_groups#index"
  match "/db/channel_groups",                                 via: :post,                                                 to: "v4/db/channel_groups#create"
  match "/db/channel_groups/:id",                             via: :delete, as: :db_channel_group_detail,                 to: "v4/db/channel_groups#destroy"
  match "/db/channel_groups/:id",                             via: :patch,                                                to: "v4/db/channel_groups#update"
  match "/db/channel_groups/:id/edit",                        via: :get,    as: :db_edit_channel_group,                   to: "v4/db/channel_groups#edit"
  match "/db/channel_groups/:id/publishing",                  via: :delete, as: :db_channel_group_publishing,             to: "v4/db/channel_group_publishings#destroy"
  match "/db/channel_groups/:id/publishing",                  via: :post,                                                 to: "v4/db/channel_group_publishings#create"
  match "/db/channel_groups/new",                             via: :get,    as: :db_new_channel_group,                    to: "v4/db/channel_groups#new"
  match "/db/channels",                                       via: :get,    as: :db_channel_list,                         to: "v4/db/channels#index"
  match "/db/channels",                                       via: :post,                                                 to: "v4/db/channels#create"
  match "/db/channels/:id",                                   via: :delete, as: :db_channel_detail,                       to: "v4/db/channels#destroy"
  match "/db/channels/:id",                                   via: :patch,                                                to: "v4/db/channels#update"
  match "/db/channels/:id/edit",                              via: :get,    as: :db_edit_channel,                         to: "v4/db/channels#edit"
  match "/db/channels/:id/publishing",                        via: :delete, as: :db_channel_publishing,                   to: "v4/db/channel_publishings#destroy"
  match "/db/channels/:id/publishing",                        via: :post,                                                 to: "v4/db/channel_publishings#create"
  match "/db/channels/new",                                   via: :get,    as: :db_new_channel,                          to: "v4/db/channels#new"
  match "/db/characters",                                     via: :get,    as: :db_character_list,                       to: "v4/db/characters#index"
  match "/db/characters",                                     via: :post,                                                 to: "v4/db/characters#create"
  match "/db/characters/:id",                                 via: :delete, as: :db_character_detail,                     to: "v4/db/characters#destroy"
  match "/db/characters/:id",                                 via: :patch,                                                to: "v4/db/characters#update"
  match "/db/characters/:id/edit",                            via: :get,    as: :db_edit_character,                       to: "v4/db/characters#edit"
  match "/db/characters/:id/publishing",                      via: :delete, as: :db_character_publishing,                 to: "v4/db/character_publishings#destroy"
  match "/db/characters/:id/publishing",                      via: :post,                                                 to: "v4/db/character_publishings#create"
  match "/db/characters/new",                                 via: :get,    as: :db_new_character,                        to: "v4/db/characters#new"
  match "/db/episodes/:id",                                   via: :delete, as: :db_episode_detail,                       to: "v4/db/episodes#destroy"
  match "/db/episodes/:id",                                   via: :patch,                                                to: "v4/db/episodes#update"
  match "/db/episodes/:id/edit",                              via: :get,    as: :db_edit_episode,                         to: "v4/db/episodes#edit"
  match "/db/episodes/:id/publishing",                        via: :delete, as: :db_episode_publishing,                   to: "v4/db/episode_publishings#destroy"
  match "/db/episodes/:id/publishing",                        via: :post,                                                 to: "v4/db/episode_publishings#create"
  match "/db/organizations",                                  via: :get,    as: :db_organization_list,                    to: "v4/db/organizations#index"
  match "/db/organizations",                                  via: :post,                                                 to: "v4/db/organizations#create"
  match "/db/organizations/:id",                              via: :delete, as: :db_organization_detail,                  to: "v4/db/organizations#destroy"
  match "/db/organizations/:id",                              via: :patch,                                                to: "v4/db/organizations#update"
  match "/db/organizations/:id/edit",                         via: :get,    as: :db_edit_organization,                    to: "v4/db/organizations#edit"
  match "/db/organizations/:id/publishing",                   via: :delete, as: :db_organization_publishing,              to: "v4/db/organization_publishings#destroy"
  match "/db/organizations/:id/publishing",                   via: :post,                                                 to: "v4/db/organization_publishings#create"
  match "/db/organizations/new",                              via: :get,    as: :db_new_organization,                     to: "v4/db/organizations#new"
  match "/db/people",                                         via: :get,    as: :db_person_list,                          to: "v4/db/people#index"
  match "/db/people",                                         via: :post,                                                 to: "v4/db/people#create"
  match "/db/people/:id",                                     via: :delete, as: :db_person_detail,                        to: "v4/db/people#destroy"
  match "/db/people/:id",                                     via: :patch,                                                to: "v4/db/people#update"
  match "/db/people/:id/edit",                                via: :get,    as: :db_edit_person,                          to: "v4/db/people#edit"
  match "/db/people/:id/publishing",                          via: :delete, as: :db_person_publishing,                    to: "v4/db/person_publishings#destroy"
  match "/db/people/:id/publishing",                          via: :post,                                                 to: "v4/db/person_publishings#create"
  match "/db/people/new",                                     via: :get,    as: :db_new_person,                           to: "v4/db/people#new"
  match "/db/programs/:id",                                   via: :delete, as: :db_program_detail,                       to: "v4/db/programs#destroy"
  match "/db/programs/:id",                                   via: :patch,                                                to: "v4/db/programs#update"
  match "/db/programs/:id/edit",                              via: :get,    as: :db_edit_program,                         to: "v4/db/programs#edit"
  match "/db/programs/:id/publishing",                        via: :delete, as: :db_program_publishing,                   to: "v4/db/program_publishings#destroy"
  match "/db/programs/:id/publishing",                        via: :post,                                                 to: "v4/db/program_publishings#create"
  match "/db/search",                                         via: :get,    as: :db_search,                               to: "v4/db/searches#show"
  match "/db/series_works/:id",                               via: :delete, as: :db_series_work_detail,                   to: "v4/db/series_works#destroy"
  match "/db/series_works/:id",                               via: :patch,                                                to: "v4/db/series_works#update"
  match "/db/series_works/:id/edit",                          via: :get,    as: :db_edit_series_work,                     to: "v4/db/series_works#edit"
  match "/db/series_works/:id/publishing",                    via: :delete, as: :db_series_work_publishing,               to: "v4/db/series_work_publishings#destroy"
  match "/db/series_works/:id/publishing",                    via: :post,                                                 to: "v4/db/series_work_publishings#create"
  match "/db/series",                                         via: :get,    as: :db_series_list,                          to: "v4/db/series#index"
  match "/db/series",                                         via: :post,                                                 to: "v4/db/series#create"
  match "/db/series/:id",                                     via: :delete, as: :db_series_detail,                        to: "v4/db/series#destroy"
  match "/db/series/:id",                                     via: :patch,                                                to: "v4/db/series#update"
  match "/db/series/:id/edit",                                via: :get,    as: :db_edit_series,                          to: "v4/db/series#edit"
  match "/db/series/:id/publishing",                          via: :delete, as: :db_series_publishing,                    to: "v4/db/series_publishings#destroy"
  match "/db/series/:id/publishing",                          via: :post,                                                 to: "v4/db/series_publishings#create"
  match "/db/series/:series_id/series_works",                 via: :get,    as: :db_series_work_list,                     to: "v4/db/series_works#index"
  match "/db/series/:series_id/series_works",                 via: :post,                                                 to: "v4/db/series_works#create"
  match "/db/series/:series_id/series_works/new",             via: :get,    as: :db_new_series_work,                      to: "v4/db/series_works#new"
  match "/db/series/new",                                     via: :get,    as: :db_new_series,                           to: "v4/db/series#new"
  match "/db/slots/:id",                                      via: :delete, as: :db_slot_detail,                          to: "v4/db/slots#destroy"
  match "/db/slots/:id",                                      via: :patch,                                                to: "v4/db/slots#update"
  match "/db/slots/:id/edit",                                 via: :get,    as: :db_edit_slot,                            to: "v4/db/slots#edit"
  match "/db/slots/:id/publishing",                           via: :delete, as: :db_slot_publishing,                      to: "v4/db/slot_publishings#destroy"
  match "/db/slots/:id/publishing",                           via: :post,                                                 to: "v4/db/slot_publishings#create"
  match "/db/staffs/:id",                                     via: :delete, as: :db_staff_detail,                         to: "v4/db/staffs#destroy"
  match "/db/staffs/:id",                                     via: :patch,                                                to: "v4/db/staffs#update"
  match "/db/staffs/:id/edit",                                via: :get,    as: :db_edit_staff,                           to: "v4/db/staffs#edit"
  match "/db/staffs/:id/publishing",                          via: :delete, as: :db_staff_publishing,                     to: "v4/db/staff_publishings#destroy"
  match "/db/staffs/:id/publishing",                          via: :post,                                                 to: "v4/db/staff_publishings#create"
  match "/db/trailers/:id",                                   via: :delete, as: :db_trailer_detail,                       to: "v4/db/trailers#destroy"
  match "/db/trailers/:id",                                   via: :patch,                                                to: "v4/db/trailers#update"
  match "/db/trailers/:id/edit",                              via: :get,    as: :db_edit_trailer,                         to: "v4/db/trailers#edit"
  match "/db/trailers/:id/publishing",                        via: :delete, as: :db_trailer_publishing,                   to: "v4/db/trailer_publishings#destroy"
  match "/db/trailers/:id/publishing",                        via: :post,                                                 to: "v4/db/trailer_publishings#create"
  match "/db/works",                                          via: :get,    as: :db_work_list,                            to: "v4/db/works#index"
  match "/db/works",                                          via: :post,                                                 to: "v4/db/works#create"
  match "/db/works/:id",                                      via: :delete, as: :db_work_detail,                          to: "v4/db/works#destroy"
  match "/db/works/:id",                                      via: :patch,                                                to: "v4/db/works#update"
  match "/db/works/:id/edit",                                 via: :get,    as: :db_edit_work,                            to: "v4/db/works#edit"
  match "/db/works/:id/publishing",                           via: :delete, as: :db_work_publishing,                      to: "v4/db/work_publishings#destroy"
  match "/db/works/:id/publishing",                           via: :post,                                                 to: "v4/db/work_publishings#create"
  match "/db/works/:work_id/casts",                           via: :get,    as: :db_cast_list,                            to: "v4/db/casts#index"
  match "/db/works/:work_id/casts",                           via: :post,                                                 to: "v4/db/casts#create"
  match "/db/works/:work_id/casts/new",                       via: :get,    as: :db_new_cast,                             to: "v4/db/casts#new"
  match "/db/works/:work_id/episodes",                        via: :get,    as: :db_episode_list,                         to: "v4/db/episodes#index"
  match "/db/works/:work_id/episodes",                        via: :post,                                                 to: "v4/db/episodes#create"
  match "/db/works/:work_id/episodes/new",                    via: :get,    as: :db_new_episode,                          to: "v4/db/episodes#new"
  match "/db/works/:work_id/image",                           via: :get,    as: :db_work_image_detail,                    to: "v4/db/work_images#show"
  match "/db/works/:work_id/image",                           via: :patch,                                                to: "v4/db/work_images#update"
  match "/db/works/:work_id/image",                           via: :post,                                                 to: "v4/db/work_images#create"
  match "/db/works/:work_id/programs",                        via: :get,    as: :db_program_list,                         to: "v4/db/programs#index"
  match "/db/works/:work_id/programs",                        via: :post,                                                 to: "v4/db/programs#create"
  match "/db/works/:work_id/programs/new",                    via: :get,    as: :db_new_program,                          to: "v4/db/programs#new"
  match "/db/works/:work_id/slots",                           via: :get,    as: :db_slot_list,                            to: "v4/db/slots#index"
  match "/db/works/:work_id/slots",                           via: :post,                                                 to: "v4/db/slots#create"
  match "/db/works/:work_id/slots/new",                       via: :get,    as: :db_new_slot,                             to: "v4/db/slots#new"
  match "/db/works/:work_id/staffs",                          via: :get,    as: :db_staff_list,                           to: "v4/db/staffs#index"
  match "/db/works/:work_id/staffs",                          via: :post,                                                 to: "v4/db/staffs#create"
  match "/db/works/:work_id/staffs/new",                      via: :get,    as: :db_new_staff,                            to: "v4/db/staffs#new"
  match "/db/works/:work_id/trailers",                        via: :get,    as: :db_trailer_list,                         to: "v4/db/trailers#index"
  match "/db/works/:work_id/trailers",                        via: :post,                                                 to: "v4/db/trailers#create"
  match "/db/works/:work_id/trailers/new",                    via: :get,    as: :db_new_trailer,                          to: "v4/db/trailers#new"
  match "/db/works/new",                                      via: :get,    as: :db_new_work,                             to: "v4/db/works#new"
  match "/dummy_image",                                       via: :get,                                                  to: "application#dummy_image" if Rails.env.test?
  match "/episode_records",                                   via: :patch,  as: :episode_record_mutation,                 to: "v4/episode_records#update"
  match "/episodes/:episode_id/records",                      via: :post,   as: :episode_record_list,                     to: "episode_records#create"
  match "/faq",                                               via: :get,    as: :faq,                                     to: "v6/faqs#show"
  match "/forum",                                             via: :get,    as: :forum,                                   to: "v3/forum/home#show"
  match "/forum/categories/:category_id",                     via: :get,    as: :forum_category,                          to: "v3/forum/categories#show",           category_id: /[a-z_]+/
  match "/forum/posts",                                       via: :post,   as: :forum_post_list,                         to: "v3/forum/posts#create"
  match "/forum/posts/:post_id",                              via: :get,    as: :forum_post,                              to: "v3/forum/posts#show",                post_id: ID_FORMAT
  match "/forum/posts/:post_id",                              via: :patch,                                                to: "v3/forum/posts#update",              post_id: ID_FORMAT
  match "/forum/posts/:post_id/comments",                     via: :post,   as: :forum_comment_list,                      to: "v3/forum/comments#create",           post_id: ID_FORMAT
  match "/forum/posts/:post_id/comments/:comment_id",         via: :patch,  as: :forum_comment,                           to: "v3/forum/comments#update",           post_id: ID_FORMAT, comment_id: ID_FORMAT
  match "/forum/posts/:post_id/comments/:comment_id/edit",    via: :get,    as: :forum_edit_comment,                      to: "v3/forum/comments#edit",             post_id: ID_FORMAT, comment_id: ID_FORMAT
  match "/forum/posts/:post_id/edit",                         via: :get,    as: :forum_edit_post,                         to: "v3/forum/posts#edit",                post_id: ID_FORMAT
  match "/forum/posts/new",                                   via: :get,    as: :forum_new_post,                          to: "v3/forum/posts#new"
  match "/fragment/@:username/records/:record_id/edit",       via: :get,    as: :fragment_edit_record,                    to: "v6/fragment/records#edit",           username: USERNAME_FORMAT
  match "/fragment/@:username/tracking_heatmap",              via: :get,    as: :fragment_tracking_heatmap,               to: "v6/fragment/tracking_heatmaps#show", username: USERNAME_FORMAT
  match "/fragment/activity_groups/:activity_group_id/items", via: :get,    as: :fragment_activity_item_list,             to: "v6/fragment/activity_items#index"
  match "/fragment/episodes/:episode_id/records",             via: :get,    as: :fragment_episode_record_list,            to: "v6/fragment/episode_records#index"
  match "/fragment/receive_channel_buttons",                  via: :get,    as: :fragment_receive_channel_button_list,    to: "v6/fragment/receive_channel_buttons#index"
  match "/fragment/trackable_anime/:anime_id",                via: :get,    as: :fragment_trackable_anime,                to: "v6/fragment/trackable_anime#show"
  match "/fragment/trackable_episodes",                       via: :get,    as: :fragment_trackable_episode_list,         to: "v6/fragment/trackable_episodes#index"
  match "/fragment/trackable_episodes/:episode_id",           via: :get,    as: :fragment_trackable_episode,              to: "v6/fragment/trackable_episodes#show"
  match "/friends",                                           via: :get,    as: :friend_list,                             to: "v3/friends#index"
  match "/legal",                                             via: :get,    as: :legal,                                   to: "v6/pages#legal"
  match "/notifications",                                     via: :get,    as: :notification_list,                       to: "v3/notifications#index"
  match "/organizations/:organization_id",                    via: :get,    as: :organization,                            to: "v3/organizations#show"
  match "/organizations/:organization_id/fans",               via: :get,    as: :organization_fan_list,                   to: "v3/organization_fans#index"
  match "/people/:person_id",                                 via: :get,    as: :person,                                  to: "v3/people#show"
  match "/people/:person_id/fans",                            via: :get,    as: :person_fan_list,                         to: "v3/person_fans#index"
  match "/privacy",                                           via: :get,    as: :privacy,                                 to: "v6/pages#privacy"
  match "/registrations/new",                                 via: :get,    as: :new_registration,                        to: "v6/registrations#new"
  match "/search",                                            via: :get,    as: :search,                                  to: "v3/searches#show"
  match "/settings",                                          via: :get,    as: :setting_list,                            to: "v3/settings#index"
  match "/settings/account",                                  via: :get,    as: :settings_account,                        to: "v3/settings/accounts#show"
  match "/settings/account",                                  via: :patch,                                                to: "v3/settings/accounts#update"
  match "/settings/apps",                                     via: :get,    as: :settings_app_list,                       to: "v3/settings/apps#index"
  match "/settings/apps/:app_id/revoke",                      via: :patch,  as: :settings_revoke_app,                     to: "v3/settings/apps#revoke"
  match "/settings/email_notification",                       via: :get,    as: :settings_email_notification,             to: "v3/settings/email_notifications#show"
  match "/settings/email_notification",                       via: :patch,                                                to: "v3/settings/email_notifications#update"
  match "/settings/email_notification/unsubscribe",           via: :get,    as: :settings_unsubscribe_email_notification, to: "v3/settings/email_notifications#unsubscribe"
  match "/settings/muted_users",                              via: :get,    as: :settings_muted_user_list,                to: "v3/settings/muted_users#index"
  match "/settings/muted_users/:mute_user_id",                via: :delete, as: :settings_muted_user,                     to: "v6/settings/muted_users#destroy"
  match "/settings/options",                                  via: :get,    as: :settings_option_list,                    to: "v3/settings/options#index"
  match "/settings/options",                                  via: :patch,                                                to: "v3/settings/options#update"
  match "/settings/password",                                 via: :patch,  as: :settings_password,                       to: "v3/settings/passwords#update"
  match "/settings/profile",                                  via: :get,    as: :settings_profile,                        to: "v3/settings/profiles#show"
  match "/settings/profile",                                  via: :patch,                                                to: "v3/settings/profiles#update"
  match "/settings/providers",                                via: :get,    as: :settings_provider_list,                  to: "v3/settings/providers#index"
  match "/settings/providers/:provider_id",                   via: :delete, as: :settings_provider,                       to: "v3/settings/providers#destroy"
  match "/settings/tokens",                                   via: :post,   as: :settings_token_list,                     to: "v3/settings/tokens#create"
  match "/settings/tokens/:token_id",                         via: :delete, as: :settings_token,                          to: "v3/settings/tokens#destroy"
  match "/settings/tokens/:token_id",                         via: :patch,                                                to: "v3/settings/tokens#update"
  match "/settings/tokens/:token_id/edit",                    via: :get,    as: :settings_edit_token,                     to: "v3/settings/tokens#edit"
  match "/settings/tokens/new",                               via: :get,    as: :settings_new_token,                      to: "v3/settings/tokens#new"
  match "/settings/user",                                     via: :delete, as: :settings_user,                           to: "v3/settings/users#destroy"
  match "/sign_in",                                           via: :get,    as: :new_user_session,                        to: "v6/sign_in#new" # for Devise
  match "/sign_in",                                           via: :get,    as: :sign_in,                                 to: "v6/sign_in#new"
  match "/sign_in/callback",                                  via: :get,    as: :sign_in_callback,                        to: "v6/sign_in_callbacks#show"
  match "/sign_up",                                           via: :get,    as: :sign_up,                                 to: "v6/sign_up#new"
  match "/supporters",                                        via: :get,    as: :supporters,                              to: "v3/supporters#show"
  match "/terms",                                             via: :get,    as: :terms,                                   to: "v6/pages#terms"
  match "/track",                                             via: :get,    as: :track,                                   to: "tracks#show"
  match "/userland",                                          via: :get,    as: :userland,                                to: "v3/userland/home#show"
  match "/userland/projects",                                 via: :post,   as: :userland_project_list,                   to: "v3/userland/projects#create"
  match "/userland/projects/:project_id",                     via: :delete, as: :userland_project,                        to: "v3/userland/projects#destroy",   project_id: ID_FORMAT
  match "/userland/projects/:project_id",                     via: :get,                                                  to: "v3/userland/projects#show",   project_id: ID_FORMAT
  match "/userland/projects/:project_id",                     via: :patch,                                                to: "v3/userland/projects#update", project_id: ID_FORMAT
  match "/userland/projects/:project_id/edit",                via: :get,    as: :userland_edit_project,                   to: "v3/userland/projects#edit",   project_id: ID_FORMAT
  match "/userland/projects/new",                             via: :get,    as: :userland_new_project,                    to: "v3/userland/projects#new"
  match "/work_display_option",                               via: :get,    as: :work_display_option,                     to: "v3/work_display_options#show"
  match "/works/:anime_id",                                   via: :get,    as: :anime,                                   to: "v4/works#show",               anime_id: ID_FORMAT
  match "/works/:anime_id/episodes",                          via: :get,    as: :episode_list,                            to: "v4/episodes#index",           anime_id: ID_FORMAT
  match "/works/:anime_id/episodes/:episode_id",              via: :get,    as: :episode,                                 to: "v6/episodes#show",            anime_id: ID_FORMAT
  match "/works/:anime_id/records",                           via: :get,    as: :anime_record_list,                       to: "v4/anime_records#index",      anime_id: ID_FORMAT
  match "/works/:anime_id/records",                           via: :post,                                                 to: "v4/anime_records#create",     anime_id: ID_FORMAT
  match "/works/:slug",                                       via: :get,    as: :seasonal_anime_list,                     to: "v3/works#season",             slug: /[0-9]{4}-(all|spring|summer|autumn|winter)/
  match "/works/newest",                                      via: :get,    as: :newest_anime_list,                       to: "v3/works#newest"
  match "/works/popular",                                     via: :get,    as: :popular_anime_list,                      to: "v3/works#popular"
  # standard:enable Layout/ExtraSpacing, Layout/LineLength

  root "v6/home#show",
    constraints: Annict::RoutingConstraints::Member.new
  root "v6/welcome#show",
    constraints: Annict::RoutingConstraints::Guest.new,
    # Set :as option to avoid two routes with the same name
    as: nil
end
