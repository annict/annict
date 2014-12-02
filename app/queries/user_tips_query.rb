class UserTipsQuery
  def initialize(user)
    @user = user
  end

  def unfinished(target = :all)
    finished_tip_ids = @user.finished_tips.pluck(:tip_id)
    Tip.where.not(id: finished_tip_ids).order(:id)
  end
end
