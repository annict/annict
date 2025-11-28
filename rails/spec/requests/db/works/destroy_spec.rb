# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/works/:id", type: :request do
  it "ユーザーがサインインしていないとき、ログインページにリダイレクトすること" do
    work = create(:work, :not_deleted)

    expect(Work.count).to eq(1)

    delete "/db/works/#{work.id}"
    work.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")

    expect(Work.count).to eq(1)
  end

  it "エディター権限を持たないユーザーがサインインしているとき、アクセスが拒否されること" do
    user = create(:registered_user)
    work = create(:work, :not_deleted)
    login_as(user, scope: :user)

    expect(Work.count).to eq(1)

    delete "/db/works/#{work.id}"
    work.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(Work.count).to eq(1)
  end

  it "エディター権限を持つユーザーがサインインしているとき、アクセスが拒否されること" do
    user = create(:registered_user, :with_editor_role)
    work = create(:work, :not_deleted)
    login_as(user, scope: :user)

    expect(Work.count).to eq(1)

    delete "/db/works/#{work.id}"
    work.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(Work.count).to eq(1)
  end

  it "管理者権限を持つユーザーがサインインしているとき、作品を論理削除できること" do
    user = create(:registered_user, :with_admin_role)
    work = create(:work, :not_deleted)
    login_as(user, scope: :user)

    expect(Work.count).to eq(1)

    delete "/db/works/#{work.id}"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")

    expect(Work.count).to eq(0)
  end

  it "存在しない作品IDを指定したとき、404エラーになること" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    expect { delete "/db/works/non-existent-id" }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "既に削除済みの作品を削除しようとしたとき、404エラーになること" do
    user = create(:registered_user, :with_admin_role)
    work = create(:work, :not_deleted)
    login_as(user, scope: :user)

    # 先に削除する
    work.destroy!

    expect { delete "/db/works/#{work.id}" }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
