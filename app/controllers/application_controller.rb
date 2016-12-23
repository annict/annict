# frozen_string_literal: true

class ApplicationController < ActionController::Base
  include ControllerCommon
  include PageCategoryHelper
  include ViewSelector
  include FlashMessage

  # Prevent CSRF attacks by raising an exception.
  # For APIs, you may want to use :null_session instead.
  protect_from_forgery with: :exception

  helper_method :client_uuid, :gon

  before_action :load_data_into_gon
  before_action :set_search_params
  before_action :load_new_user

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

  def load_work
    @work = Work.published.find(params[:work_id])
  end

  def load_episode
    @episode = @work.episodes.published.find(params[:episode_id])
  end

  def set_search_params
    @search = SearchService.new(params[:q])
  end

  def load_new_user
    return if user_signed_in?
    @new_user = User.new_with_session({}, session)
  end
end
