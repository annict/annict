# frozen_string_literal: true

class ApplicationController < ActionController::Base
  include Pundit

  include ControllerCommon
  include Analyzable
  include LogrageSetting
  include Gonable
  include PageCategoryHelper
  include ViewSelector
  include FlashMessage

  # Prevent CSRF attacks by raising an exception.
  # For APIs, you may want to use :null_session instead.
  protect_from_forgery with: :exception, prepend: true

  helper_method :gon

  before_action :redirect_if_unexpected_subdomain
  before_action :switch_locale
  before_action :set_search_params
  before_action :load_new_user
  before_action :load_data_into_gon

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

  def after_sign_out_path_for(_resource_or_scope)
    root_path
  end

  def localable_resources(resources)
    if user_signed_in?
      resources.with_locale(current_user.allowed_locales)
    elsif !user_signed_in? && locale_en?
      resources.with_locale(:en)
    elsif !user_signed_in? && locale_ja?
      resources.with_locale(:ja)
    else
      resources
    end
  end

  def load_work
    @work = Work.published.find(params[:work_id])
  end

  def load_character
    @character = Character.published.find(params[:character_id])
  end

  def load_episode
    @episode = @work.episodes.published.find(params[:episode_id])
  end
end
