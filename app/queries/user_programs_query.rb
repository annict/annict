class UserProgramsQuery
  attr_reader :user

  def initialize(user)
    @user = user
  end

  # チェックインしていないエピソードと紐づく番組情報を返す
  def unchecked
    works = user.works.wanna_watch_and_watching
    channel_works = user.channel_works.where(work: works)

    Program.where(id: unchecked_program_ids(channel_works))
  end

  private

  def unchecked_episodes(work)
    UserEpisodesQuery.new(user).unchecked(work)
  end

  def unchecked_program_ids(channel_works)
    program_ids = []

    channel_works.each do |cw|
      episode_ids = unchecked_episodes(cw.work).pluck(:id)
      conditions = { channel_id: cw.channel_id, episode_id: episode_ids }
      program_ids << Program.where(conditions).pluck(:id)
    end

    program_ids.flatten
  end
end
