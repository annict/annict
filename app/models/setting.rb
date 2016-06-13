# frozen_string_literal: true
# == Schema Information
#
# Table name: settings
#
#  id                       :integer          not null, primary key
#  user_id                  :integer          not null
#  hide_checkin_comment     :boolean          default(TRUE), not null
#  created_at               :datetime         not null
#  updated_at               :datetime         not null
#  share_record_to_twitter  :boolean          default(FALSE)
#  share_record_to_facebook :boolean          default(FALSE)
#  programs_sort            :string           default(""), not null
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
end
