# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/works/:work_id/image", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    work = FactoryBot.create(:work)

    get db_work_image_detail_path(work)

    expect(response).to redirect_to(new_user_session_path)
  end

  it "作品が削除されているとき、404エラーになること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work, deleted_at: Time.current)
    login_as(user, scope: :user)

    get db_work_image_detail_path(work)

    expect(response.status).to eq(404)
  end

  it "画像が設定されていないとき、200ステータスでフォームが表示されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    get db_work_image_detail_path(work)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(work.title)
  end

  it "画像が設定されているとき、200ステータスで画像情報が表示されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    # image_dataがnullにできないため、最小限のデータを設定
    work_image = WorkImage.new(
      work:,
      user:,
      copyright: "© Test Copyright"
    )
    # image_dataを直接設定
    work_image.image_data = "{}"
    work_image.save!
    login_as(user, scope: :user)

    get db_work_image_detail_path(work)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(work.title)
    expect(response.body).to include("© Test Copyright")
  end
end
