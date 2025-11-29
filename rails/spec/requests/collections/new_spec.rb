# typed: false
# frozen_string_literal: true

RSpec.describe "GET /collections/new", type: :request do
  it "ログインしているとき、新規コレクション作成ページが正常に表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/collections/new"

    expect(response.status).to eq(200)
  end

  it "ログインしているとき、フォームが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/collections/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("forms_collection_form")
  end

  it "ログインしているとき、コレクション名の入力フィールドが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/collections/new"

    expect(response.status).to eq(200)
    expect(response.body).to include('name="forms_collection_form[name]"')
  end

  it "ログインしているとき、説明の入力フィールドが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/collections/new"

    expect(response.status).to eq(200)
    expect(response.body).to include('name="forms_collection_form[description]"')
  end

  it "ログインしていないとき、認証エラーが発生すること" do
    expect {
      get "/collections/new"
    }.to raise_error(NoMethodError, /undefined method.*profile.*for nil/)
  end
end
