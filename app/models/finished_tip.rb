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
# Indexes
#
#  index_finished_tips_on_user_id_and_tip_id  (user_id,tip_id) UNIQUE
#

class FinishedTip < ActiveRecord::Base
  belongs_to :tip
end
