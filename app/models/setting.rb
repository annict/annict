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
#  programs_sort_type            :string           default(""), not null
#  display_option_work_list      :string           default("list"), not null
#  display_option_user_work_list :string           default("list"), not null
#
# Indexes
#
#  index_settings_on_user_id  (user_id)
#

class Setting < ActiveRecord::Base
  extend Enumerize

  belongs_to :user

  enumerize :programs_sort_type,
    in: %i(started_at_asc started_at_desc),
    default: :started_at_desc

  enumerize :display_option_work_list,
    in: %i(list grid grid_small),
    default: :list

  enumerize :display_option_user_work_list,
    in: %i(list grid grid_small),
    default: :list
end
