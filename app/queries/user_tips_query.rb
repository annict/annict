class UserTipsQuery
  def initialize(user)
    @user = user
  end

  def unfinished(target = :all)
    values = (target == :all) ? Tip.target.values : target
    finished_tip_ids = @user.finished_tips.pluck(:tip_id)
    Tip.where.not(id: finished_tip_ids).order(:id)
  end
end
