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
#  index_receptions_on_user_id_and_channel_id  (user_id,channel_id) UNIQUE
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
