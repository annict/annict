# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /db/works/:work_id/image", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    work = FactoryBot.create(:work)
    work_image_params = {
      copyright: "© Example Studio"
    }

    patch "/db/works/#{work.id}/image", params: {work_image: work_image_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "編集者権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    work_image = WorkImage.create!(
      work: work,
      user: user,
      copyright: "© Original Studio",
      image_data: {
        "id" => "test.jpg",
        "storage" => "cache",
        "metadata" => {"size" => 12345, "filename" => "test.jpg", "mime_type" => "image/jpeg"}
      }.to_json
    )
    work_image_params = {
      copyright: "© Updated Studio"
    }

    login_as(user, scope: :user)

    patch "/db/works/#{work.id}/image", params: {work_image: work_image_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    work_image.reload
    expect(work_image.copyright).to eq("© Original Studio")
  end

  it "編集者権限を持つユーザーがログインしているとき、画像情報を更新できること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    work = FactoryBot.create(:work)
    work_image = WorkImage.create!(
      work: work,
      user: user,
      copyright: "© Original Studio",
      image_data: {
        "id" => "test.jpg",
        "storage" => "cache",
        "metadata" => {"size" => 12345, "filename" => "test.jpg", "mime_type" => "image/jpeg"}
      }.to_json
    )
    work_image_params = {
      copyright: "© Updated Studio 2024"
    }

    login_as(user, scope: :user)

    patch "/db/works/#{work.id}/image", params: {work_image: work_image_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("保存しました")

    work_image.reload
    expect(work_image.copyright).to eq("© Updated Studio 2024")
    expect(work_image.user).to eq(user)
  end

  it "編集者権限を持つユーザーがログインしているとき、バリデーションエラーで更新に失敗すること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    work = FactoryBot.create(:work)
    work_image = WorkImage.create!(
      work: work,
      user: user,
      copyright: "© Original Studio",
      image_data: {
        "id" => "test.jpg",
        "storage" => "cache",
        "metadata" => {"size" => 12345, "filename" => "test.jpg", "mime_type" => "image/jpeg"}
      }.to_json
    )
    work_image_params = {
      copyright: "" # 必須フィールドを空にしてバリデーションエラーを発生させる
    }

    login_as(user, scope: :user)

    patch "/db/works/#{work.id}/image", params: {work_image: work_image_params}

    expect(response.status).to eq(200)
    expect(response.body).to include("著作者情報を入力してください")

    work_image.reload
    expect(work_image.copyright).to eq("© Original Studio")
  end

  it "管理者権限を持つユーザーがログインしているとき、画像情報を更新できること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    work = FactoryBot.create(:work)
    work_image = WorkImage.create!(
      work: work,
      user: user,
      copyright: "© Original Studio",
      image_data: {
        "id" => "test.jpg",
        "storage" => "cache",
        "metadata" => {"size" => 12345, "filename" => "test.jpg", "mime_type" => "image/jpeg"}
      }.to_json
    )
    work_image_params = {
      copyright: "© Admin Updated Studio"
    }

    login_as(user, scope: :user)

    patch "/db/works/#{work.id}/image", params: {work_image: work_image_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("保存しました")

    work_image.reload
    expect(work_image.copyright).to eq("© Admin Updated Studio")
    expect(work_image.user).to eq(user)
  end

  it "画像がまだ存在しない作品に対してPATCHリクエストを送ったとき、404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    work = FactoryBot.create(:work)
    work_image_params = {
      copyright: "© New Studio"
    }

    login_as(user, scope: :user)

    patch "/db/works/#{work.id}/image", params: {work_image: work_image_params}

    expect(response.status).to eq(404)
  end

  it "削除された作品に対してPATCHリクエストを送ったとき、404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    work = FactoryBot.create(:work, :deleted)
    work_image_params = {
      copyright: "© Updated Studio"
    }

    login_as(user, scope: :user)

    patch "/db/works/#{work.id}/image", params: {work_image: work_image_params}

    expect(response.status).to eq(404)
  end
end
