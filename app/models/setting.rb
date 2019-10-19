# frozen_string_literal: true
# == Schema Information
#
# Table name: settings
#
#  id                            :integer          not null, primary key
#  user_id                       :integer          not null
#  hide_record_comment           :boolean          default(TRUE), not null
#  created_at                    :datetime         not null
#  updated_at                    :datetime         not null
#  share_record_to_twitter       :boolean          default(FALSE)
#  share_record_to_facebook      :boolean          default(FALSE)
#  slots_sort_type            :string           default(NULL), not null
#  display_option_work_list      :string           default("list_detailed"), not null
#  display_option_user_work_list :string           default("grid_detailed"), not null
#  records_sort_type             :string           default("created_at_desc"), not null
#  display_option_record_list    :string           default("all_comments"), not null
#  share_review_to_twitter       :boolean          default(FALSE), not null
#  share_review_to_facebook      :boolean          default(FALSE), not null
#  hide_supporter_badge          :boolean          default(FALSE), not null
#  share_status_to_twitter       :boolean          default(FALSE), not null
#  share_status_to_facebook      :boolean          default(FALSE), not null
#  privacy_policy_agreed         :boolean          default(FALSE), not null
#
# Indexes
#
#  index_settings_on_user_id  (user_id)
#

class Setting < ApplicationRecord
  extend Enumerize

  belongs_to :user

  enumerize :slots_sort_type,
    in: %i(started_at_asc started_at_desc),
    default: :started_at_desc

  enumerize :records_sort_type,
    in: %i(
      likes_count_desc
      rating_state_desc
      rating_state_asc
      created_at_desc
      created_at_asc
    ),
    default: :created_at_desc

  enumerize :display_option_work_list,
    in: %i(
      grid
      grid_small
      list_detailed
    ),
    default: :list_detailed

  enumerize :display_option_user_work_list,
    in: %i(
      grid
      grid_detailed
      grid_small
    ),
    default: :grid_detailed

  enumerize :display_option_record_list,
    in: %i(
      all_comments
      friend_comments
      my_episode_records
    ),
    default: :all_comments
end
