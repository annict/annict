# == Schema Information
#
# Table name: receptions
#
#  id         :integer          not null, primary key
#  user_id    :integer          not null
#  channel_id :integer          not null
#  created_at :datetime
#  updated_at :datetime
#
# Indexes
#
#  receptions_channel_id_idx          (channel_id)
#  receptions_user_id_channel_id_key  (user_id,channel_id) UNIQUE
#  receptions_user_id_idx             (user_id)
#

class Reception < ActiveRecord::Base
  belongs_to :channel
  belongs_to :user

  after_create :finish_tips


  private

  def finish_tips
    if user.first_reception?(self)
      tip = Tip.find_by(partial_name: 'channel')
      user.finish_tip!(tip)
    end
  end
end
