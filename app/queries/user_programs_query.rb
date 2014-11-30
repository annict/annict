class UserProgramsQuery
  def initialize(user)
    @user = user
  end

  # チェックインしていないエピソードと紐づく番組情報を返す
  def unchecked
    works = @user.works.wanna_watch_and_watching
    channel_works = @user.channel_works.where(work: works)

    conditions = channel_works.map do |cw|
      "(work_id = #{cw.work_id} and channel_id = #{cw.channel_id})"
    end

    Program.where(conditions.join(' OR '))
  end
end
