# typed: false
# frozen_string_literal: true

class Setting < ApplicationRecord
  extend Enumerize

  self.ignored_columns = %w[
    display_option_record_list display_option_user_work_list
    share_record_to_facebook share_review_to_facebook share_review_to_twitter
    timeline_mode
  ]

  belongs_to :user

  enumerize :slots_sort_type,
    in: %i[started_at_asc started_at_desc],
    default: :started_at_desc

  enumerize :records_sort_type,
    in: %i[
      likes_count_desc
      rating_state_desc
      rating_state_asc
      created_at_desc
      created_at_asc
    ],
    default: :created_at_desc

  enumerize :display_option_work_list,
    in: %i[
      grid
      grid_small
      list_detailed
    ],
    default: :list_detailed

  enumerize :display_option_user_work_list,
    in: %i[
      grid
      grid_detailed
      grid_small
    ],
    default: :grid_detailed

  enumerize :display_option_record_list,
    in: %i[
      all_comments
      friend_comments
      my_episode_records
    ],
    default: :all_comments
end
