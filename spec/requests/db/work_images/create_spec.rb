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

    # 実際の画像アップロードのテストはImageUploadableやアップローダーで行う
    # ここではコントローラーの動作確認に必要な最小限のパラメータを送る
    expect {
      post "/db/works/#{work.id}/image", params: {
        work_image: {
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

  it "著作権情報が255文字を超える場合、エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    expect {
      post "/db/works/#{work.id}/image", params: {
        work_image: {
          copyright: "a" * 256
        }
      }
    }.not_to change(WorkImage, :count)

    expect(response).to have_http_status(:ok)
    # copyrightが255文字以上のときのエラーメッセージを確認
    expect(response.body).to include("is too long")
  end
end
