# frozen_string_literal: true

# standard:disable Layout/ExtraSpacing, Layout/LineLength
match "/forum",                                          via: :get,   as: :forum,              to: "v3/forum/home#show"
match "/forum/categories/:category_id",                  via: :get,   as: :forum_category,     to: "v3/forum/categories#show", category_id: /[a-z_]+/
match "/forum/posts",                                    via: :post,  as: :forum_post_list,    to: "v3/forum/posts#create"
match "/forum/posts/:post_id",                           via: :get,   as: :forum_post,         to: "v3/forum/posts#show",      post_id: /[0-9]+/
match "/forum/posts/:post_id",                           via: :patch,                          to: "v3/forum/posts#update",    post_id: /[0-9]+/
match "/forum/posts/:post_id/comments",                  via: :post,  as: :forum_comment_list, to: "v3/forum/comments#create", post_id: /[0-9]+/
match "/forum/posts/:post_id/comments/:comment_id",      via: :patch, as: :forum_comment,      to: "v3/forum/comments#update", post_id: /[0-9]+/, comment_id: /[0-9]+/
match "/forum/posts/:post_id/comments/:comment_id/edit", via: :get,   as: :forum_edit_comment, to: "v3/forum/comments#edit",   post_id: /[0-9]+/, comment_id: /[0-9]+/
match "/forum/posts/:post_id/edit",                      via: :get,   as: :forum_edit_post,    to: "v3/forum/posts#edit",      post_id: /[0-9]+/
match "/forum/posts/new",                                via: :get,   as: :forum_new_post,     to: "v3/forum/posts#new"
# standard:enable Layout/ExtraSpacing, Layout/LineLength
