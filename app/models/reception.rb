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
