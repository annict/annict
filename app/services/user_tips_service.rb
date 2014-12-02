class UserTipsService
  def initialize(user)
    @user = user
  end

  def finish!(partial_name)
    tip = Tip.find_by(partial_name: partial_name.to_s)
    @user.finished_tips.where(tip: tip).first_or_create!
  end
end
