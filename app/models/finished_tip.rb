# == Schema Information
#
# Table name: finished_tips
#
#  id         :bigint           not null, primary key
#  created_at :datetime         not null
#  updated_at :datetime         not null
#  tip_id     :bigint           not null
#  user_id    :bigint           not null
#
# Indexes
#
#  index_finished_tips_on_user_id_and_tip_id  (user_id,tip_id) UNIQUE
#
# Foreign Keys
#
#  finished_tips_tip_id_fk   (tip_id => tips.id) ON DELETE => cascade
#  finished_tips_user_id_fk  (user_id => users.id) ON DELETE => cascade
#

class FinishedTip < ApplicationRecord
  belongs_to :tip
end
