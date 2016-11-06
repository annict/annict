# frozen_string_literal: true

class UserTipsService
  def initialize(user)
    @user = user
  end

  def finish!(slug)
    tip = Tip.find_by(slug: slug)
    @user.finished_tips.where(tip: tip).first_or_create!
  end
end
