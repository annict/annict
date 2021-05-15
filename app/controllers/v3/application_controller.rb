# frozen_string_literal: true

module V3
  class ApplicationController < ActionController::Base
    include Pundit

    include V3::ControllerCommon
    include V3::Analyzable
    include V3::LogrageSetting
    include V3::Gonable
    include V3::ViewSelector
    include V3::FlashMessage
    include V6::SentryLoadable
    include V6::Localizable
    include V6::PageCategorizable
    include V6::KeywordSearchable

    layout "v3/default"

    # Prevent CSRF attacks by raising an exception.
    # For APIs, you may want to use :null_session instead.
    protect_from_forgery with: :exception, prepend: true

    helper_method :gon, :locale_ja?, :locale_en?, :local_url, :page_category

    before_action :load_new_user
    before_action :store_data_into_gon

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
end
