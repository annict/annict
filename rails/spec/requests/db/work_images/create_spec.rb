# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/works/:work_id/image", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    work = FactoryBot.create(:work)

    post "/db/works/#{work.id}/image"

    expect(response).to redirect_to(new_user_session_path)
  end

  it "作品が削除されているとき、404エラーになること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work, deleted_at: Time.current)
    login_as(user, scope: :user)

    expect {
      post "/db/works/#{work.id}/image"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "権限がないユーザーの場合、403エラーになること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    post "/db/works/#{work.id}/image", params: {
      work_image: {
        copyright: "© Example"
      }
    }

    expect(response).to redirect_to(db_root_path)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "編集者権限を持つユーザーが有効なパラメータで画像を作成できること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    # 画像ファイルをアップロードするテスト
    # fixture_file_uploadを使用して実際のファイルアップロードをシミュレート
    image_file = fixture_file_upload("test_image.jpg", "image/jpeg")

    expect {
      post "/db/works/#{work.id}/image", params: {
        work_image: {
          image: image_file,
          copyright: "© Example"
        }
      }
    }.to change(WorkImage, :count).by(1)

    expect(response).to redirect_to(db_work_image_detail_path(work))
    expect(flash[:notice]).to eq(I18n.t("messages.work_images.saved"))

    work_image = WorkImage.last
    expect(work_image.work).to eq(work)
    expect(work_image.user).to eq(user)
    expect(work_image.copyright).to eq("© Example")
  end

  it "無効なパラメータの場合、画像が作成されず画面が再描画されること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    expect {
      post "/db/works/#{work.id}/image", params: {
        work_image: {
          copyright: ""
        }
      }
    }.not_to change(WorkImage, :count)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("著作者情報を入力してください")
  end

  it "著作者情報が256文字でも保存できること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    copyright = "a" * 256
    image_file = fixture_file_upload("test_image.jpg", "image/jpeg")

    expect {
      post "/db/works/#{work.id}/image", params: {
        work_image: {
          image: image_file,
          copyright: copyright
        }
      }
    }.to change(WorkImage, :count).by(1)

    expect(response).to redirect_to(db_work_image_detail_path(work))
    expect(WorkImage.last.copyright).to eq(copyright)
  end
end
