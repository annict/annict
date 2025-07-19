# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /sign_out", type: :request do
  it "ログイン済みユーザーが正常にログアウトできること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    delete "/sign_out"

    expect(response).to redirect_to("/")
    expect(flash[:notice]).to include("ログアウトしました")
  end

  it "ログインしていないユーザーがアクセスしてもエラーにならないこと" do
    delete "/sign_out"

    expect(response).to redirect_to("/")
  end

  it "ログアウト後にセッションが削除されていること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    delete "/sign_out"

    # ログイン済みユーザーのみアクセス可能なページにアクセスしてリダイレクトされることを確認
    get "/settings/account"
    expect(response).to redirect_to("/sign_in")
  end
end
