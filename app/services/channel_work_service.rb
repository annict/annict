class ChannelWorkService
  def initialize(user)
    @user = user
  end

  def channel_work(work)
    @user.channel_works.find_by(anime_id: work.id)
  end

  # チャンネルと紐付いていない作品と最速放送チャンネルとを紐付ける
  def create(work)
    if channel_work(work).blank?
      channel = @user.channels.fastest(work)

      @user.channel_works.create(anime: work, channel: channel) if channel.present?
    end
  end

  # まだ放送予定が存在しなかったために `channel_works` に
  # どのチャンネルで見るかの情報が保存されなかった作品にcronで対処するためのメソッド
  def update
    @user.works.wanna_watch_and_watching.each do |work|
      create(work) if channel_work(work).blank?
    end
  end

  def delete(work)
    channel_work(work).destroy if channel_work(work).present?
  end
end
