class ApplicationController < ActionController::Base
  include FlashMessage

  # Prevent CSRF attacks by raising an exception.
  # For APIs, you may want to use :null_session instead.
  protect_from_forgery with: :exception

  before_filter :store_user_info


  # テスト実行時にDragonflyでアップロードした画像を読み込むときに呼ばれるアクション
  # 画像サーバはこのRailsアプリから切り離しているので、CircleCI等でテストを実行するときは
  # このダミーのアクションを画像だと思って呼ぶ
  def dummy_image
    # テストを実行するときは画像は表示されていなくて良いので、ただ200を返す
    head 200
  end

  private

  def after_sign_in_path_for(resource)
    root_path
  end

  def after_sign_out_path_for(resource_or_scope)
    root_path
  end

  def set_work
    @work = Work.find(params[:work_id])
  end

  def set_episode
    @episode = @work.episodes.find(params[:episode_id])
  end

  def set_checkin
    @checkin = @episode.checkins.find(params[:checkin_id])
  end

  def store_user_info
    if user_signed_in?
      user_info = {
        shareRecordToTwitter: current_user.setting.share_record_to_twitter?,
        shareRecordToFacebook: current_user.setting.share_record_to_facebook?,
        sharableToTwitter: current_user.shareable_to?(:twitter),
        sharableToFacebook: current_user.shareable_to?(:facebook)
      }

      gon.push(user_info)
    end
  end
end
