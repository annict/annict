# frozen_string_literal: true

class ApplicationController < ActionController::Base
  include Pundit::Authorization

  include BasicAuthenticatable
  include ControllerCommon
  include Analyzable
  include ViewSelector
  include SentryLoadable
  include Localizable
  include PageCategorizable
  include KeywordSearchable

  layout "main_default"

  # Prevent CSRF attacks by raising an exception.
  # For APIs, you may want to use :null_session instead.
  protect_from_forgery with: :exception, prepend: true

  around_action :set_locale

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
end
