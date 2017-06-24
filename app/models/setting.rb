# frozen_string_literal: true
# == Schema Information
#
# Table name: settings
#
#  id                            :integer          not null, primary key
#  user_id                       :integer          not null
#  hide_checkin_comment          :boolean          default(TRUE), not null
#  created_at                    :datetime         not null
#  updated_at                    :datetime         not null
#  share_record_to_twitter       :boolean          default(FALSE)
#  share_record_to_facebook      :boolean          default(FALSE)
#  programs_sort_type            :string           default(NULL), not null
#  display_option_work_list      :string           default("list"), not null
#  display_option_user_work_list :string           default("list"), not null
#  records_sort_type             :string           default("created_at_desc"), not null
#  display_option_record_list    :string           default("all_comments"), not null
#  share_review_to_twitter       :boolean          default(FALSE), not null
#  share_review_to_facebook      :boolean          default(FALSE), not null
#
# Indexes
#
#  index_settings_on_user_id  (user_id)
#

class Setting < ApplicationRecord
  extend Enumerize

  belongs_to :user

  enumerize :programs_sort_type,
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
    in: %i(list grid grid_small),
    default: :list

  enumerize :display_option_user_work_list,
    in: %i(list grid grid_small),
    default: :list

  enumerize :display_option_record_list,
    in: %i(
      all_comments
      friend_comments
      my_records
    ),
    default: :all_comments
end
