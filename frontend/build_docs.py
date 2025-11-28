import os
from pathlib import Path
from collections import defaultdict

def scan_repository(root_path, output_file="documentacao_codigo.md"):
    """
    Varre o repositório e gera documentação em Markdown
    
    Args:
        root_path: Caminho raiz do repositório
        output_file: Nome do arquivo de saída
    """
    
    # Extensões de arquivo para processar
    extensions = {'.tsx', '.ts', '.jsx', '.js', '.css', '.json', '.md', '.mjs'}
    
    # Pastas e arquivos para ignorar
    ignore_patterns = {
        'node_modules', '.next', '.git', 'dist', 'build', 
        '.env.local', 'package-lock.json', '.gitignore', 'README.md',
        'package.json', 'tsconfig.json', 'postcss.config.mjs', 
        'eslint.config.mjs', 'tailwind.config.ts', 'brazil-locations.ts'
    }
    
    files_with_content = []
    empty_files = []
    files_by_folder = defaultdict(list)
    
    # Primeira passagem: coletar informações
    for dirpath, dirnames, filenames in os.walk(root_path):
        # Filtrar pastas ignoradas
        dirnames[:] = [d for d in dirnames if d not in ignore_patterns]
        
        rel_dir = os.path.relpath(dirpath, root_path)
        
        for filename in filenames:
            # Ignorar arquivos específicos
            if filename in ignore_patterns:
                continue
                
            file_ext = Path(filename).suffix
            
            # Processar apenas extensões relevantes
            if file_ext not in extensions:
                continue
            
            full_path = os.path.join(dirpath, filename)
            rel_path = os.path.relpath(full_path, root_path)
            
            try:
                with open(full_path, 'r', encoding='utf-8') as f:
                    content = f.read().strip()
                
                file_info = {
                    'path': rel_path,
                    'full_path': full_path,
                    'content': content,
                    'size': len(content)
                }
                
                if content:
                    files_with_content.append(file_info)
                    files_by_folder[rel_dir].append(file_info)
                else:
                    empty_files.append(rel_path)
                    
            except Exception as e:
                print(f"Erro ao ler {rel_path}: {e}")
    
    # Gerar arquivo Markdown
    with open(output_file, 'w', encoding='utf-8') as md:
        md.write("# Documentação do Código Frontend\n\n")
        md.write(f"**Total de arquivos com conteúdo:** {len(files_with_content)}\n")
        md.write(f"**Total de arquivos vazios:** {len(empty_files)}\n\n")
        md.write("---\n\n")
        
        # Ordenar pastas
        sorted_folders = sorted(files_by_folder.keys())
        
        for folder in sorted_folders:
            files = sorted(files_by_folder[folder], key=lambda x: x['path'])
            
            # Cabeçalho da pasta
            md.write(f"## 📁 {folder}\n\n")
            
            for file_info in files:
                rel_path = file_info['path'].replace('\\', '/')
                content = file_info['content']
                
                md.write(f"### {rel_path}\n\n")
                md.write("```" + Path(rel_path).suffix[1:] + "\n")
                md.write(content)
                md.write("\n```\n\n")
                md.write("---\n\n")
        
        # Resumo final
        md.write("\n## 📊 Resumo Final\n\n")
        
        md.write("### ✅ Arquivos com Conteúdo\n\n")
        for file_info in sorted(files_with_content, key=lambda x: x['path']):
            path = file_info['path'].replace('\\', '/')
            size = file_info['size']
            md.write(f"- `{path}` ({size} caracteres)\n")
        
        md.write(f"\n**Total:** {len(files_with_content)} arquivos\n\n")
        
        if empty_files:
            md.write("### ⚠️ Arquivos Vazios\n\n")
            for empty_path in sorted(empty_files):
                path = empty_path.replace('\\', '/')
                md.write(f"- `{path}`\n")
            md.write(f"\n**Total:** {len(empty_files)} arquivos\n\n")
        
        # Estatísticas por tipo de arquivo
        md.write("### 📈 Estatísticas por Tipo\n\n")
        ext_stats = defaultdict(lambda: {'count': 0, 'total_size': 0})
        
        for file_info in files_with_content:
            ext = Path(file_info['path']).suffix
            ext_stats[ext]['count'] += 1
            ext_stats[ext]['total_size'] += file_info['size']
        
        for ext in sorted(ext_stats.keys()):
            count = ext_stats[ext]['count']
            size = ext_stats[ext]['total_size']
            md.write(f"- **{ext}**: {count} arquivo(s), {size:,} caracteres\n")
    
    print(f"\n✅ Documentação gerada com sucesso: {output_file}")
    print(f"📁 Arquivos com conteúdo: {len(files_with_content)}")
    print(f"⚠️  Arquivos vazios: {len(empty_files)}")

if __name__ == "__main__":
    # Executar no diretório atual
    root = "."
    
    # Ou especifique o caminho do seu projeto:
    # root = "C:/caminho/para/seu/projeto"
    
    scan_repository(root)