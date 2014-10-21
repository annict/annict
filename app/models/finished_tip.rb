# == Schema Information
#
# Table name: finished_tips
#
#  id         :integer          not null, primary key
#  user_id    :integer          not null
#  tip_id     :integer          not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#

class FinishedTip < ActiveRecord::Base
end
