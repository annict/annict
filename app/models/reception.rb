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


  def self.initial?(reception)
    count == 1 && first.id == reception.id
  end


  private

  def finish_tips
    if user.receptions.initial?(self)
      UserTipsService.new(user).finish!(:channel)
    end
  end
end
