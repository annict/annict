# frozen_string_literal: true

# == Schema Information
#
# Table name: settings
#
#  id                       :bigint           not null, primary key
#  display_option_work_list :string           default("list_detailed"), not null
#  hide_record_body         :boolean          default(TRUE), not null
#  hide_supporter_badge     :boolean          default(FALSE), not null
#  privacy_policy_agreed    :boolean          default(FALSE), not null
#  records_sort_type        :string           default("created_at_desc"), not null
#  share_record_to_twitter  :boolean          default(FALSE)
#  share_status_to_facebook :boolean          default(FALSE), not null
#  share_status_to_twitter  :boolean          default(FALSE), not null
#  slots_sort_type          :string           default(NULL), not null
#  created_at               :datetime         not null
#  updated_at               :datetime         not null
#  user_id                  :bigint           not null
#
# Indexes
#
#  index_settings_on_user_id  (user_id)
#
# Foreign Keys
#
#  fk_rails_...  (user_id => users.id)
#

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
