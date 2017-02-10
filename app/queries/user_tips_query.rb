# frozen_string_literal: true

class UserTipsQuery
  def initialize(user)
    @user = user
  end

  def unfinished
    finished_tip_ids = @user.finished_tips.pluck(:tip_id)
    Tip.where.not(id: finished_tip_ids).order(:id)
  end

  def unfinished_for_new_user
    unfinished.with_target(:new_user)
  end
end
