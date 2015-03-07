class UserEpisodesQuery
  attr_reader :user

  def initialize(user)
    @user = user
  end

  # 指定した作品の中のチェックインしていないエピソードを返す
  def unchecked(work_or_works)
    checked_episode_ids = user.checkins.where(work: work_or_works).pluck(:episode_id)
    Episode.where(work: work_or_works).where.not(id: checked_episode_ids)
  end
end
