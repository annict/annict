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
