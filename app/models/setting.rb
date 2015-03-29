# == Schema Information
#
# Table name: settings
#
#  id                   :integer          not null, primary key
#  user_id              :integer          not null
#  hide_checkin_comment :boolean          default(TRUE), not null
#  created_at           :datetime         not null
#  updated_at           :datetime         not null
#
# Indexes
#
#  index_settings_on_user_id  (user_id)
#

class Setting < ActiveRecord::Base
  belongs_to :user
end
