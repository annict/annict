# frozen_string_literal: true

class ApplicationController < ActionController::Base
  include PageCategoryHelper
  include ViewSelector
  include FlashMessage

  # Prevent CSRF attacks by raising an exception.
  # For APIs, you may want to use :null_session instead.
  protect_from_forgery with: :exception

  helper_method :client_uuid, :gon

  before_action :load_data_into_gon
  before_action :set_search_params

  # テスト実行時にDragonflyでアップロードした画像を読み込むときに呼ばれるアクション
  # 画像サーバはこのRailsアプリから切り離しているので、CircleCI等でテストを実行するときは
  # このダミーのアクションを画像だと思って呼ぶ
  def dummy_image
    # テストを実行するときは画像は表示されていなくて良いので、ただ200を返す
    head 200
  end

  private

  def after_sign_in_path_for(resource)
    path = stored_location_for(resource)
    return path if path.present?
    root_path
  end

  def after_sign_out_path_for(resource_or_scope)
    root_path
  end

  def set_work
    @work = Work.published.find(params[:work_id])
  end

  def set_episode
    @episode = @work.episodes.published.find(params[:episode_id])
  end

  def load_record
    @record = @episode.checkins.find(params[:checkin_id])
  end

  def set_search_params
    @search = SearchService.new(params[:q])
  end

  def load_data_into_gon
    data = {
      user: {
        device: browser.device.mobile? ? "mobile" : "pc",
        clientUUID: client_uuid,
        userId: user_signed_in? ? current_user.encoded_id : nil
      },
      keen: {
        projectId: ENV.fetch("KEEN_PROJECT_ID"),
        writeKey: ENV.fetch("KEEN_WRITE_KEY")
      }
    }

    if user_signed_in?
      data[:user].merge!(
        shareRecordToTwitter: current_user.setting.share_record_to_twitter?,
        shareRecordToFacebook: current_user.setting.share_record_to_facebook?,
        sharableToTwitter: current_user.shareable_to?(:twitter),
        sharableToFacebook: current_user.shareable_to?(:facebook)
      )
    end

    gon.push(data)
  end

  def load_i18n_into_gon(keys)
    gon.I18n = {}
    keys.each do |k, v|
      key = v.present? && browser.device.mobile? && v.key?(:mobile) ? v[:mobile] : k
      gon.I18n[k] = I18n.t(key)
    end
  end

  def ga_client
    @ga_client ||= Annict::Analytics::Client.new(request, current_user)
  end

  def keen_client
    @keen_client ||= Annict::Keen::Client.new(request)
  end

  def store_client_uuid
    return if cookies[:ann_client_uuid].present?

    cookies[:ann_client_uuid] = {
      value: request.uuid,
      expires: 2.year.from_now,
      domain: ".#{ENV.fetch('ANNICT_DOMAIN')}"
    }

    cookies[:ann_client_uuid]
  end

  def client_uuid
    cookies[:ann_client_uuid].presence || store_client_uuid
  end
end
