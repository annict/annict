# frozen_string_literal: true

class UserEpisodeRecordsQuery
  def call(episode_records:, user:)
    @episode_records = episode_records
    @user = user

    @episode_records = @episode_records
      .only_kept
      .preload(:record, episode: {work: :anime_image}, user: %i[profile setting gumroad_subscriber])
      .joins(:user)
      .merge(User.only_kept)
    @episode_records = @episode_records.where.not(user_id: @user.mute_users.pluck(:muted_user_id)) if @user
    @episode_records = join_likes if @user

    selects = ["episode_records.*"]
    selects << "likes.id AS user_like_id" if @user
    @episode_records.select(selects.join(", "))
  end

  private

  def join_likes
    sql = [
      "
        LEFT OUTER JOIN likes ON
          likes.recipient_type = 'EpisodeRecord' AND
          episode_records.id = likes.recipient_id AND
          likes.user_id = %s
      ", @user.id
    ]
    @episode_records.joins(Like.sanitize_sql_array(sql))
  end
end
