class UserEpisodesQuery
  def initialize(user, work_or_works)
    @user = user
    @work_or_works = work_or_works
  end

  # 指定した作品の中のチェックインしていないエピソードを返す
  def unchecked
    checked_episode_ids = @user.checkins.where(work: @work_or_works).pluck(:episode_id)
    Episode.where(work: @work_or_works).where.not(id: checked_episode_ids)
  end
end
